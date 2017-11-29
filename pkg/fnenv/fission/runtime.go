package fission

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"
	"io/ioutil"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	executor "github.com/fission/fission/executor/client"
	"github.com/sirupsen/logrus"

	"strings"

	"github.com/fission/fission/router"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var log = logrus.WithField("component", "fnenv-fission")

// FunctionEnv adapts the Fission platform to the function execution runtime.
type FunctionEnv struct {
	executor *executor.Client
	ct       *ContentTypeMapper
}

func NewFunctionEnv(executor *executor.Client) *FunctionEnv {
	return &FunctionEnv{
		executor: executor,
		ct:       &ContentTypeMapper{typedvalues.DefaultParserFormatter},
	}
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	meta := &metav1.ObjectMeta{
		Name:      spec.GetType().GetSrc(),
		UID:       k8stypes.UID(spec.GetType().GetResolved()),
		Namespace: metav1.NamespaceDefault,
	}
	ctxLog := log.WithField("fn", meta.Name)
	ctxLog.Infof("Invoking Fission function: '%v'.", meta.Name, meta.UID)
	serviceUrl, err := fe.executor.GetServiceForFunction(meta)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err":  err,
			"meta": meta,
		}).Error("Fission function failed!")
		return nil, err
	}

	url := fmt.Sprintf("http://%s", serviceUrl)

	// Map input parameters to actual Fission function parameters

	var input []byte
	mainInput, ok := spec.Inputs[types.INPUT_MAIN]
	if ok {
		input = mainInput.Value
	} else {
		mainInput, ok := spec.Inputs["body"]
		if ok {
			input = mainInput.Value
		}
	}

	r := bytes.NewReader(input)
	// TODO map other parameters as well (to params)

	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		panic(fmt.Errorf("failed to make request for '%s': %v", serviceUrl, err))
	}
	defer req.Body.Close()

	router.MetadataToHeaders(router.HEADERS_FISSION_FUNCTION_PREFIX, meta, req)

	reqContentType := ToContentType(mainInput)
	req.Header.Set("Content-Type", reqContentType)
	ctxLog.Infof("[request][Content-Type]: %v", reqContentType)
	ctxLog.Debugf("[request][body]: %v", string(input))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error for url '%s': %v", serviceUrl, err)
	}

	output := ToTypedValue(resp)
	ctxLog.Infof("[response][status]: %v", meta.Name, resp.StatusCode)
	ctxLog.Infof("[response][Content-Type]: %v ", meta.Name, resp.Header.Get("Content-Type"))
	ctxLog.Debugf("[%s][output]: %v", meta.Name, output)

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(output)
		ctxLog.Warn("[%s] Failed %v: %v", resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", resp.StatusCode),
				Message: fmt.Sprintf("%s", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
	}, nil
}

type ContentTypeMapper struct {
	parserFormatter typedvalues.ParserFormatter
}

var formatMapping = map[string]string{
	typedvalues.FORMAT_JSON: "application/json",
}

func ToContentType(val *types.TypedValue) string {
	contentType := "text/plain"
	if val == nil {
		return contentType
	}

	// Temporary solution
	if strings.HasPrefix(val.Type, "json") {
		contentType = "application/json"
	}
	return contentType
}

func ToTypedValue(resp *http.Response) *types.TypedValue {
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var i interface{} = body
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		log.Info("Assuming JSON")
		err := json.Unmarshal(body, &i)
		if err != nil {
			log.Warnf("Expected JSON response could not be parsed: %v", err)
		}
	}

	tv, err := typedvalues.Parse(i)
	if err != nil {
		panic(err)
	}
	return tv
}
