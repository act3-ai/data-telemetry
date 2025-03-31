package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/schema"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
)

func setCurrentURLParams(w http.ResponseWriter, r *http.Request, params *bottleRequestParams) *httputil.HTTPError {
	currentURL, err := url.Parse(r.Header.Get("HX-Current-URL"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusInternalServerError, `Error parsing current URL via "HX-Current-URL" request header`)
	}

	newURLQueryParams := url.Values{}
	encoder := schema.NewEncoder()
	encoder.RegisterEncoder(requestTimestamp{}, func(v reflect.Value) string {
		return fmt.Sprintf("%v", v)
	})
	err = encoder.Encode(params, newURLQueryParams)
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Error parsing current URL query parameters")
	}
	newURLQueryParams = removeUnsetURLParams(&newURLQueryParams)
	newURL := currentURL
	newURL.RawQuery = newURLQueryParams.Encode()
	w.Header().Add("HX-Push-Url", newURL.String())
	return nil
}

func removeUnsetURLParams(values *url.Values) url.Values {
	newValues := url.Values{}
	for k, vSlice := range *values {
		if values.Get(k) != "" {
			for _, v := range vSlice {
				newValues.Add(k, v)
			}
		}
	}
	return newValues
}
