package registration

//
//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"github.com/searKing/go-api-auth/server/auth"
//	"github.com/searKing/go-api-auth/server/model"
//	"io"
//	"io/ioutil"
//	"log"
//	"mime"
//	"net/http"
//	"net/url"
//)
//
//// rfc6749 2.1
//type ClientType string
//
//const (
//	ClientTypeConfidential ClientType = "confidential"
//	ClientTypePublic       ClientType = "public"
//)
//
//type ClientRegisteration struct {
//	RegisterFunc func(ctx context.Context, clientType string, clientRedirectURIs []string, w http.ResponseWriter, r *http.Request) (pass bool)
//}
//
//func parseClientInfoFromHeaderField(r *http.Request) string {
//	// rfc6750 2.1
//	// Authorization Request Header Field
//	return r.Header.Get("Client-Type")
//}
//
//func parseCredentialsFromFormEncodedBodyParameter(r *http.Request) string {
//	defer r.Body.Close()
//	// rfc6750 2.2
//	// Authorization Request Header Field
//	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
//	if err != nil {
//		return ""
//	}
//	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
//	switch content {
//	// rfc6750 2.2
//	// Form-Encoded Body Parameter
//	case "application/x-www-form-urlencoded":
//		if r.Method == http.MethodGet {
//			break
//		}
//		vals, err := url.ParseQuery(string(body))
//		if err != nil {
//			return ""
//		}
//		return vals.Get("client_type")
//	}
//	return ""
//}
//
//func parseCredentialsFromURIQueryParameter(r *http.Request) string {
//	// rfc6750 2.3
//	// URI Query Parameter
//	vars := r.URL.Query()
//	accessTokens, ok := vars["client_type"]
//	if !ok || len(accessTokens) == 0 {
//		return ""
//	}
//	return accessTokens[0]
//}
//func (c *ClientRegisteration) RegisterFuncHandler(ctx context.Context) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
//	}
//}
//
//func (c *ClientRegisteration) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) (pass bool) {
//	if c.RegisterFunc != nil {
//		return c.RegisterFunc(ctx, w, r)
//	}
//	return true
//}
//
//func Register(clientId string, clientType ClientType, handler http.Handler) http.Handler {
//
//	url := "http://localhost:8080/security_city/register"
//	info := &auth.RegisterReq{
//		AppId: appId,
//	}
//	infos, err := json.Marshal(info)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	payload := bytes.NewBuffer(infos)
//
//	req, _ := http.NewRequest("POST", url, payload)
//
//	req.Header.Add("Content-Type", "application/json")
//
//	res, _ := http.DefaultClient.Do(req)
//
//	defer res.Body.Close()
//	body, _ := ioutil.ReadAll(res.Body)
//
//	registerResp := auth.RegisterResp{}
//	err = json.Unmarshal(body, &registerResp)
//	if err != nil {
//		log.Fatal(err)
//	}
//	return model.User{
//		AppId:     registerResp.AppId,
//		AppKey:    registerResp.AppKey,
//		AppSecret: registerResp.AppSecret,
//	}
//}
