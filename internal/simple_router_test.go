package simple_router

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testRouterSuite struct {
	suite.Suite
	rtDef         RouteDefinition
	frontendProxy *httptest.Server
}

func TestRouterTest(t *testing.T) {
	suite.Run(t, &testRouterSuite{})
}

func (suite *testRouterSuite) TestBasic() {
	target1 := createTestTarget()
	defer target1.Close()

	target2 := createTestTarget()
	defer target2.Close()

	target3 := createTestTarget()
	defer target3.Close()

	suite.Equal(0, target1.callCount)
	suite.Equal(0, target2.callCount)
	suite.Equal(0, target3.callCount)

	suite.rtDef = RouteDefinition{
		RouteBySelector: RouteMap{
			"k.[l]+.+": target1.URL,
			"k.*t":     target2.URL,
			"e.+a":     target3.URL,
		},
		Selector: RouteSelector{
			Name: "x-user",
			Type: SelectorTypeHeader,
		},
	}
	router, routerCreErr := NewRouter(suite.rtDef, CreateRootLogger())
	suite.NoError(routerCreErr)
	suite.frontendProxy = httptest.NewServer(&router)
	defer suite.frontendProxy.Close()

	suite.testWithUser("kabat")
	suite.Equal(0, target1.callCount)
	suite.Equal(1, target2.callCount)
	suite.Equal(0, target3.callCount)

	suite.testWithUser("kasdft")
	suite.Equal(0, target1.callCount)
	suite.Equal(2, target2.callCount)
	suite.Equal(0, target3.callCount)

	suite.testWithUser("kalap")
	suite.Equal(1, target1.callCount)
	suite.Equal(2, target2.callCount)
	suite.Equal(0, target3.callCount)

	suite.testWithUser("kabat")
	suite.Equal(1, target1.callCount)
	suite.Equal(3, target2.callCount)
	suite.Equal(0, target3.callCount)

	suite.testWithUser("ebola")
	suite.Equal(1, target1.callCount)
	suite.Equal(3, target2.callCount)
	suite.Equal(1, target3.callCount)
}

func (suite *testRouterSuite) TestPlainProxyKeyMatching() {
	suite.rtDef = RouteDefinition{
		RouteBySelector: RouteMap{
			"kalap": "csuka",
			"kabat": "ponty",
		},
		Selector: RouteSelector{
			Name: "x-user",
			Type: SelectorTypeHeader,
		},
	}
	router, routerCreErr := NewRouter(suite.rtDef, CreateRootLogger())
	suite.NoError(routerCreErr)

	matchingProxyKeys := router.getMatchingProxyKeys("kalap")
	suite.Equal([]string{"kalap"}, matchingProxyKeys)
}

func (suite *testRouterSuite) TestPatternedProxyKeyMatching() {
	suite.rtDef = RouteDefinition{
		RouteBySelector: RouteMap{
			"k.[l]+.+": "csuka",
			"ka.at":    "ponty",
		},
		Selector: RouteSelector{
			Name: "x-user",
			Type: SelectorTypeHeader,
		},
	}
	router, routerCreErr := NewRouter(suite.rtDef, CreateRootLogger())
	suite.NoError(routerCreErr)

	matchingProxyKeys := router.getMatchingProxyKeys("kalap")
	suite.Equal([]string{"k.[l]+.+"}, matchingProxyKeys)
}

func (suite *testRouterSuite) TestWithCookie() {
	target1 := createTestTarget()
	defer target1.Close()

	target2 := createTestTarget()
	defer target2.Close()

	suite.Equal(0, target1.callCount)
	suite.Equal(0, target2.callCount)

	suite.rtDef = RouteDefinition{
		RouteBySelector: RouteMap{
			"k.[l]+.+": target1.URL,
			"k.*t":     target2.URL,
		},
		Selector: RouteSelector{
			Name: "x-user-hint",
			Type: SelectorTypeCookie,
		},
	}
	router, routerCreErr := NewRouter(suite.rtDef, CreateRootLogger())
	suite.NoError(routerCreErr)
	suite.frontendProxy = httptest.NewServer(&router)
	defer suite.frontendProxy.Close()

	suite.testWithUser("kabat")
	suite.Equal(0, target1.callCount)
	suite.Equal(1, target2.callCount)
}

func (suite *testRouterSuite) testWithUser(user string) {
	resp, err := suite.doRequest(suite.frontendProxy.URL, user)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *testRouterSuite) doRequest(target string, user string) (*http.Response, error) {
	req := &http.Request{}
	parsedUrl, urlParseErr := url.Parse(target)
	if urlParseErr != nil {
		panic(urlParseErr)
	}
	req.URL = parsedUrl
	req.Header = make(http.Header)
	switch suite.rtDef.Selector.Type {
	case SelectorTypeHeader:
		req.Header[suite.rtDef.Selector.Name] = []string{user}
	case SelectorTypeCookie:
		cookie := &http.Cookie{
			Name:  suite.rtDef.Selector.Name,
			Value: user,
		}
		req.AddCookie(cookie)
	}
	return http.DefaultClient.Do(req)
}

type testTarget struct {
	*httptest.Server
	callCount int
}

func createTestTarget() *testTarget {
	testTrgt := &testTarget{}
	testTrgt.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testTrgt.callCount++
	}))
	return testTrgt
}
