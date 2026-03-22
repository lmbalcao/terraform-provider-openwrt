package lucirpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	humanReadableCommitChanges = "commit changes"
	humanReadableCreateSection = "create section"
	humanReadableDeleteSection = "delete section"
	humanReadableGetSection    = "get section"
	humanReadableLogin         = "login"
	humanReadableShowChanges   = "show changes"
	humanReadableUpdateSection = "update section"

	methodAdd     = "add"
	methodCall    = "call"
	methodChanges = "changes"
	methodCommit  = "commit"
	methodDelete  = "delete"
	methodGet     = "get"
	methodGetAll  = "get_all"
	methodLogin   = "login"
	methodSection = "section"
	methodSet     = "set"
	methodTSet    = "tset"

	pathAuth = "/cgi-bin/luci/rpc/auth"
	pathUBus = "/cgi-bin/luci/admin/ubus"
	pathUCI  = "/cgi-bin/luci/rpc/uci"

	queryKeyAuth = "auth"

	ubusAnonymousSession = "00000000000000000000000000000000"
	ubusJSONRPCVersion   = "2.0"
	ubusObjectSession    = "session"
	ubusObjectUCI        = "uci"
	ubusResultSuccess    = 0
)

type jsonRPCInvoker interface {
	Invoke(context.Context, string, jsonRPCRequestBody) (*json.RawMessage, error)
}

type Client struct {
	jsonRPCClientUCI jsonRPCInvoker
}

func (c *Client) CommitChanges(
	ctx context.Context,
	config string,
) (bool, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableCommitChanges, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodCommit,
		Params: []json.RawMessage{
			marshalledConfig,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableCommitChanges,
		requestBody,
	)
	if err != nil {
		return false, fmt.Errorf("unable to %s: %w", humanReadableCommitChanges, err)
	}

	var result bool
	if responseBody == nil {
		return false, nil
	}

	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return false, fmt.Errorf("unable to parse %s response: %w", humanReadableCommitChanges, err)
	}

	return result, nil
}

func (c *Client) CreateSection(
	ctx context.Context,
	config string,
	sectionType string,
	section string,
	options Options,
) (bool, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableCreateSection, err)
	}

	marshalledSectionType, err := json.Marshal(sectionType)
	if err != nil {
		return false, fmt.Errorf("unable to serialize sectionType %q for %s: %w", sectionType, humanReadableCreateSection, err)
	}

	marshalledSection, err := json.Marshal(section)
	if err != nil {
		return false, fmt.Errorf("unable to serialize section %q for %s: %w", section, humanReadableCreateSection, err)
	}

	marshalledOptions, err := json.Marshal(options)
	if err != nil {
		return false, fmt.Errorf("unable to serialize options %q for %s: %w", options, humanReadableCreateSection, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodSection,
		Params: []json.RawMessage{
			marshalledConfig,
			marshalledSectionType,
			marshalledSection,
			marshalledOptions,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableCreateSection,
		requestBody,
	)
	if err != nil {
		return false, fmt.Errorf("unable to %s: %w", humanReadableCreateSection, err)
	}

	var result bool
	if responseBody == nil {
		return false, nil
	}

	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return false, fmt.Errorf("unable to parse %s response: %w", humanReadableCreateSection, err)
	}

	if !result {
		return false, fmt.Errorf("unable to %s: it is not clear why this happened", humanReadableCreateSection)
	}

	result, err = c.CommitChanges(
		ctx,
		config,
	)
	if err != nil {
		return false, fmt.Errorf("was able to %s, but could not %s: %w", humanReadableCreateSection, humanReadableCommitChanges, err)
	}

	return result, nil
}

func (c *Client) DeleteSection(
	ctx context.Context,
	config string,
	section string,
) (bool, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableDeleteSection, err)
	}

	marshalledSection, err := json.Marshal(section)
	if err != nil {
		return false, fmt.Errorf("unable to serialize section %q for %s: %w", section, humanReadableDeleteSection, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodDelete,
		Params: []json.RawMessage{
			marshalledConfig,
			marshalledSection,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableDeleteSection,
		requestBody,
	)
	if err != nil {
		return false, fmt.Errorf("unable to %s: %w", humanReadableDeleteSection, err)
	}

	var result bool
	if responseBody == nil {
		return false, nil
	}

	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return false, fmt.Errorf("unable to parse %s response: %w", humanReadableDeleteSection, err)
	}

	if !result {
		return false, fmt.Errorf("unable to %s: it is not clear why this happened", humanReadableDeleteSection)
	}

	result, err = c.CommitChanges(
		ctx,
		config,
	)
	if err != nil {
		return false, fmt.Errorf("was able to %s, but could not %s: %w", humanReadableDeleteSection, humanReadableCommitChanges, err)
	}

	return result, nil
}

func (c *Client) GetSection(
	ctx context.Context,
	config string,
	section string,
) (Options, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableGetSection, err)
	}

	marshalledSection, err := json.Marshal(section)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize section %q for %s: %w", section, humanReadableGetSection, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodGetAll,
		Params: []json.RawMessage{
			marshalledConfig,
			marshalledSection,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableGetSection,
		requestBody,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to %s: %w", humanReadableGetSection, err)
	}

	if responseBody == nil {
		return nil, fmt.Errorf("could not find section %s.%s", config, section)
	}

	var unknownResult any
	err = json.Unmarshal(*responseBody, &unknownResult)
	if err != nil {
		return nil, fmt.Errorf("unable to determine type of %s response: %w", humanReadableGetSection, err)
	}

	_, ok := unknownResult.([]any)
	if ok {
		return nil, fmt.Errorf("incorrect config (%q) and/or section (%q): result from LuCI: %s", config, section, *responseBody)
	}

	var result Options
	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableGetSection, err)
	}

	return result, nil
}

func (c *Client) ShowChanges(
	ctx context.Context,
	config string,
) ([][]string, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableShowChanges, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodChanges,
		Params: []json.RawMessage{
			marshalledConfig,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableShowChanges,
		requestBody,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to %s: %w", humanReadableShowChanges, err)
	}

	result := [][]string{}
	if responseBody == nil {
		return result, nil
	}

	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableShowChanges, err)
	}

	return result, nil
}

func (c *Client) UpdateSection(
	ctx context.Context,
	config string,
	section string,
	options Options,
) (bool, error) {
	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("unable to serialize config %q for %s: %w", config, humanReadableUpdateSection, err)
	}

	marshalledSection, err := json.Marshal(section)
	if err != nil {
		return false, fmt.Errorf("unable to serialize section %q for %s: %w", section, humanReadableUpdateSection, err)
	}

	marshalledOptions, err := json.Marshal(options)
	if err != nil {
		return false, fmt.Errorf("unable to serialize options %q for %s: %w", options, humanReadableCreateSection, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodTSet,
		Params: []json.RawMessage{
			marshalledConfig,
			marshalledSection,
			marshalledOptions,
		},
	}
	responseBody, err := c.jsonRPCClientUCI.Invoke(
		ctx,
		humanReadableUpdateSection,
		requestBody,
	)
	if err != nil {
		return false, fmt.Errorf("unable to %s: %w", humanReadableUpdateSection, err)
	}

	var result bool
	if responseBody == nil {
		return false, nil
	}

	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return false, fmt.Errorf("unable to parse %s response: %w", humanReadableUpdateSection, err)
	}

	if !result {
		return false, fmt.Errorf("unable to %s: it is not clear why this happened", humanReadableUpdateSection)
	}

	result, err = c.CommitChanges(
		ctx,
		config,
	)
	if err != nil {
		return false, fmt.Errorf("was able to %s, but could not %s: %w", humanReadableUpdateSection, humanReadableCommitChanges, err)
	}

	return result, nil
}

func NewClient(
	ctx context.Context,
	scheme string,
	hostname string,
	port uint16,
	username string,
	password string,
) (*Client, error) {
	host := hostname
	if port != 0 {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	httpClient := &http.Client{}
	jsonRPCClientUCI, err := newLegacyUCIClient(
		ctx,
		*httpClient,
		host,
		scheme,
		username,
		password,
	)
	if err == nil {
		return &Client{jsonRPCClientUCI: jsonRPCClientUCI}, nil
	}

	if !isHTTPStatusError(err, http.StatusNotFound) {
		return nil, err
	}

	jsonRPCClientUCI, err = newUBusUCIClient(
		ctx,
		*httpClient,
		host,
		scheme,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}

	return &Client{jsonRPCClientUCI: jsonRPCClientUCI}, nil
}

func newLegacyUCIClient(
	ctx context.Context,
	httpClient http.Client,
	host string,
	scheme string,
	username string,
	password string,
) (jsonRPCInvoker, error) {
	address := url.URL{
		Host:   host,
		Path:   pathAuth,
		Scheme: scheme,
	}
	marshalledUsername, err := json.Marshal(username)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize username for %s: %w", humanReadableLogin, err)
	}

	marshalledPassword, err := json.Marshal(password)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize password for %s: %w", humanReadableLogin, err)
	}

	requestBody := jsonRPCRequestBody{
		Method: methodLogin,
		Params: []json.RawMessage{
			marshalledUsername,
			marshalledPassword,
		},
	}
	jsonRPCClient := jsonRPCNewClient(
		httpClient,
		address,
	)
	responseBody, err := jsonRPCClient.InvokeNotNull(
		ctx,
		humanReadableLogin,
		requestBody,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to %s: %w", humanReadableLogin, err)
	}

	var authToken string
	err = json.Unmarshal(responseBody, &authToken)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableLogin, err)
	}

	query := url.Values{}
	query.Add(queryKeyAuth, authToken)
	addressUCI := url.URL{
		Host:     host,
		Path:     pathUCI,
		RawQuery: query.Encode(),
		Scheme:   scheme,
	}
	return jsonRPCNewClient(
		httpClient,
		addressUCI,
	), nil
}

func newUBusUCIClient(
	ctx context.Context,
	httpClient http.Client,
	host string,
	scheme string,
	username string,
	password string,
) (jsonRPCInvoker, error) {
	address := url.URL{
		Host:   host,
		Path:   pathUBus,
		Scheme: scheme,
	}
	jsonRPCClient := ubusRPCNewClient(
		httpClient,
		address,
	)
	authToken, err := jsonRPCClient.Login(
		ctx,
		username,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to %s: %w", humanReadableLogin, err)
	}

	return ubusUCIClient{
		authToken: authToken,
		client:    jsonRPCClient,
	}, nil
}

type unexpectedStatusError struct {
	humanReadableMethod string
	status              string
	statusCode          int
}

func (e unexpectedStatusError) Error() string {
	return fmt.Sprintf("expected %s to respond with a 200: got %s", e.humanReadableMethod, e.status)
}

func isHTTPStatusError(err error, statusCode int) bool {
	var target unexpectedStatusError
	return errors.As(err, &target) && target.statusCode == statusCode
}

type jsonRPCClient struct {
	address url.URL
	client  http.Client
}

func (c jsonRPCClient) InvokeNotNull(
	ctx context.Context,
	humanReadableMethod string,
	requestBody jsonRPCRequestBody,
) (json.RawMessage, error) {
	result, err := c.Invoke(
		ctx,
		humanReadableMethod,
		requestBody,
	)
	if err != nil {
		return json.RawMessage{}, err
	}

	if result == nil {
		return nil, fmt.Errorf("invalid %s response: expected either an error or a result, got neither", humanReadableMethod)
	}

	return *result, nil
}

func (c jsonRPCClient) Invoke(
	ctx context.Context,
	humanReadableMethod string,
	requestBody jsonRPCRequestBody,
) (*json.RawMessage, error) {
	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	err := encoder.Encode(requestBody)
	if err != nil {
		return nil, fmt.Errorf("problem encoding %s request: %w", humanReadableMethod, err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.address.String(),
		&buffer,
	)
	if err != nil {
		return nil, fmt.Errorf("problem creating %s request: %w", humanReadableMethod, err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("problem sending request to %s: %w", humanReadableMethod, err)
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, unexpectedStatusError{
			humanReadableMethod: humanReadableMethod,
			status:              response.Status,
			statusCode:          response.StatusCode,
		}
	}

	var responseBody jsonRPCResponseBody
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableMethod, err)
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("%s error: %s", humanReadableMethod, *responseBody.Error)
	}

	return responseBody.Result, nil
}

func jsonRPCNewClient(
	httpClient http.Client,
	address url.URL,
) jsonRPCClient {
	return jsonRPCClient{
		address: address,
		client:  httpClient,
	}
}

type ubusUCIClient struct {
	authToken string
	client    ubusRPCClient
}

func (c ubusUCIClient) Invoke(
	ctx context.Context,
	humanReadableMethod string,
	requestBody jsonRPCRequestBody,
) (*json.RawMessage, error) {
	switch requestBody.Method {
	case methodCommit:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		_, err = c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodCommit,
			c.authToken,
			map[string]any{"config": config},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCBoolResult(true), nil

	case methodChanges:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		responseBody, err := c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodChanges,
			c.authToken,
			map[string]any{"config": config},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCObjectField(responseBody, humanReadableMethod, "changes")

	case methodDelete:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		section, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 1, "section")
		if err != nil {
			return nil, err
		}
		_, err = c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodDelete,
			c.authToken,
			map[string]any{"config": config, "section": section},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCBoolResult(true), nil

	case methodGetAll:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		section, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 1, "section")
		if err != nil {
			return nil, err
		}
		responseBody, err := c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodGet,
			c.authToken,
			map[string]any{"config": config, "section": section},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCObjectField(responseBody, humanReadableMethod, "values")

	case methodSection:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		sectionType, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 1, "section type")
		if err != nil {
			return nil, err
		}
		section, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 2, "section")
		if err != nil {
			return nil, err
		}
		options, err := jsonRPCRequestOptionsParam(requestBody, humanReadableMethod, 3)
		if err != nil {
			return nil, err
		}
		_, err = c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodAdd,
			c.authToken,
			map[string]any{"config": config, "type": sectionType, "name": section, "values": options},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCBoolResult(true), nil

	case methodTSet:
		config, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 0, "config")
		if err != nil {
			return nil, err
		}
		section, err := jsonRPCRequestStringParam(requestBody, humanReadableMethod, 1, "section")
		if err != nil {
			return nil, err
		}
		options, err := jsonRPCRequestOptionsParam(requestBody, humanReadableMethod, 2)
		if err != nil {
			return nil, err
		}
		_, err = c.client.Invoke(
			ctx,
			humanReadableMethod,
			ubusObjectUCI,
			methodSet,
			c.authToken,
			map[string]any{"config": config, "section": section, "values": options},
		)
		if err != nil {
			return nil, err
		}
		return jsonRPCBoolResult(true), nil

	default:
		return nil, fmt.Errorf("unsupported %s request %q", humanReadableMethod, requestBody.Method)
	}
}

func jsonRPCBoolResult(value bool) *json.RawMessage {
	result := json.RawMessage("false")
	if value {
		result = json.RawMessage("true")
	}
	return &result
}

func jsonRPCObjectField(
	responseBody *json.RawMessage,
	humanReadableMethod string,
	field string,
) (*json.RawMessage, error) {
	if responseBody == nil {
		return nil, nil
	}

	var object map[string]json.RawMessage
	err := json.Unmarshal(*responseBody, &object)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableMethod, err)
	}

	fieldValue, ok := object[field]
	if !ok {
		return nil, nil
	}

	return &fieldValue, nil
}

func jsonRPCRequestStringParam(
	requestBody jsonRPCRequestBody,
	humanReadableMethod string,
	index int,
	name string,
) (string, error) {
	if len(requestBody.Params) <= index {
		return "", fmt.Errorf("invalid %s request: missing %s", humanReadableMethod, name)
	}

	var result string
	err := json.Unmarshal(requestBody.Params[index], &result)
	if err != nil {
		return "", fmt.Errorf("unable to parse %s %s: %w", humanReadableMethod, name, err)
	}

	return result, nil
}

func jsonRPCRequestOptionsParam(
	requestBody jsonRPCRequestBody,
	humanReadableMethod string,
	index int,
) (Options, error) {
	if len(requestBody.Params) <= index {
		return nil, fmt.Errorf("invalid %s request: missing options", humanReadableMethod)
	}

	var result Options
	err := json.Unmarshal(requestBody.Params[index], &result)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s options: %w", humanReadableMethod, err)
	}

	return result, nil
}

type ubusRPCClient struct {
	address url.URL
	client  http.Client
}

func (c ubusRPCClient) Login(
	ctx context.Context,
	username string,
	password string,
) (string, error) {
	responseBody, err := c.Invoke(
		ctx,
		humanReadableLogin,
		ubusObjectSession,
		methodLogin,
		ubusAnonymousSession,
		map[string]any{"username": username, "password": password},
	)
	if err != nil {
		return "", err
	}

	if responseBody == nil {
		return "", fmt.Errorf("invalid %s response: expected either an error or a result, got neither", humanReadableLogin)
	}

	var result ubusLoginResponseBody
	err = json.Unmarshal(*responseBody, &result)
	if err != nil {
		return "", fmt.Errorf("unable to parse %s response: %w", humanReadableLogin, err)
	}

	if result.Session == "" {
		return "", fmt.Errorf("invalid %s response: missing ubus_rpc_session", humanReadableLogin)
	}

	return result.Session, nil
}

func (c ubusRPCClient) Invoke(
	ctx context.Context,
	humanReadableMethod string,
	object string,
	method string,
	authToken string,
	arguments any,
) (*json.RawMessage, error) {
	marshalledAuthToken, err := json.Marshal(authToken)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize auth token for %s: %w", humanReadableMethod, err)
	}

	marshalledObject, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize object for %s: %w", humanReadableMethod, err)
	}

	marshalledMethod, err := json.Marshal(method)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize method for %s: %w", humanReadableMethod, err)
	}

	marshalledArguments, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize arguments for %s: %w", humanReadableMethod, err)
	}

	requestBody := ubusRPCRequestBody{
		JSONRPC: ubusJSONRPCVersion,
		ID:      1,
		Method:  methodCall,
		Params: []json.RawMessage{
			marshalledAuthToken,
			marshalledObject,
			marshalledMethod,
			marshalledArguments,
		},
	}

	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	err = encoder.Encode(requestBody)
	if err != nil {
		return nil, fmt.Errorf("problem encoding %s request: %w", humanReadableMethod, err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.address.String(),
		&buffer,
	)
	if err != nil {
		return nil, fmt.Errorf("problem creating %s request: %w", humanReadableMethod, err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("problem sending request to %s: %w", humanReadableMethod, err)
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, unexpectedStatusError{
			humanReadableMethod: humanReadableMethod,
			status:              response.Status,
			statusCode:          response.StatusCode,
		}
	}

	var responseBody ubusRPCResponseBody
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&responseBody)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableMethod, err)
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("%s error: %s", humanReadableMethod, responseBody.Error.Message)
	}

	if responseBody.Result == nil || len(responseBody.Result) == 0 {
		return nil, fmt.Errorf("invalid %s response: expected either an error or a result, got neither", humanReadableMethod)
	}

	var resultCode int
	err = json.Unmarshal(responseBody.Result[0], &resultCode)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s response: %w", humanReadableMethod, err)
	}

	if resultCode != ubusResultSuccess {
		return nil, fmt.Errorf("%s error: ubus returned code %d", humanReadableMethod, resultCode)
	}

	if len(responseBody.Result) == 1 {
		return nil, nil
	}

	return &responseBody.Result[1], nil
}

func ubusRPCNewClient(
	httpClient http.Client,
	address url.URL,
) ubusRPCClient {
	return ubusRPCClient{
		address: address,
		client:  httpClient,
	}
}

type jsonRPCRequestBody struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

type jsonRPCResponseBody struct {
	Error  *string          `json:"error"`
	Result *json.RawMessage `json:"result"`
}

type ubusLoginResponseBody struct {
	Session string `json:"ubus_rpc_session"`
}

type ubusRPCRequestBody struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      int               `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

type ubusRPCResponseBody struct {
	Error  *ubusRPCError     `json:"error"`
	Result []json.RawMessage `json:"result"`
}

type ubusRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
