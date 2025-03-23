package verniy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type queryRequest struct {
	Query     string        `json:"query"`
	Variables queryVariable `json:"variables"`
}

type queryVariable map[string]interface{}

type errorResponse struct {
	Errors []struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"errors"`
}

func (c *Client) handleError(body []byte) error {
	var e errorResponse
	if err := json.Unmarshal(body, &e); err != nil {
		return err
	}

	errMsgs := make([]string, len(e.Errors))
	for i, b := range e.Errors {
		errMsgs[i] = b.Message
	}

	return errors.New(strings.Join(errMsgs, " | "))
}

func (c *Client) post(ctx context.Context, query string, v map[string]interface{}, model interface{}) error {
	d, err := json.Marshal(queryRequest{
		Query:     query,
		Variables: v,
	})
	if err != nil {
		return err
	}

	body, code, err := c.MakeRequest(ctx, d)
	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return c.handleError(body)
	}

	if err = json.Unmarshal(body, &model); err != nil {
		return err
	}

	return nil
}

// MakeRequest to make direct HTTP request without any wrapper.
// Prepare your body request and handle the response by yourself.
//
// Use this if you want to make custom request. Also, need to read the
// docs of how to prepare the body request and read the response
// (https://anilist.github.io/ApiV2-GraphQL-Docs/).
//
// Params requestBody is your data in JSON format. So, marshal your data
// first before passing it to this function. And will return response body,
// response code, and error.
//
// Example:
//
//  query := verniy.FieldObject("query", verniy.QueryParam{
//    "$id":   "Int",
//    "$type": "MediaType",
//  }, verniy.FieldObject("Media", verniy.QueryParam{
//    "id":   "$id",
//    "type": "$type",
//  }, "id"))
//
//  body := map[string]interface{}{
//    "query": query,
//    "variables": map[string]interface{}{
//      "id":   1,
//      "type": "ANIME",
//    },
//  }
//
//  jsonBody, _ := json.Marshal(body)
//
//  data, code, err := c.MakeRequest(jsonBody)
//  if err != nil {
//    panic(err)
//  }
//
//  fmt.Println(code)
//  fmt.Println(string(data))
func (c *Client) MakeRequest(ctx context.Context, requestBody []byte) ([]byte, int, error) {
	c.Limiter.Take()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Host, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if c.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return body, resp.StatusCode, nil
}
