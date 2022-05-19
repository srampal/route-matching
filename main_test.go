package main
// operation can be "addRoute" or "routeLookup"
// If operation is "addRoute", 
//      arg1 is  the path, 
//      arg2 is the type of match ("exact" or "prefix")
//      arg3 is the destination
//      expected1 is nil
// If operation is "routeLookup"
//      arg1 is the path,
//      expected1 is the destination expected 
//      expected2 should is any errors that are expected to be returned 


import "testing"

type pathLookupsTest struct {
     operation, arg1, arg2, arg3, expected1, expected2   string 
}

var addLookupTests = []pathLookupsTest{
    pathLookupsTest{"addRoute", "/api/1", "exact", "service-1","","", },
    pathLookupsTest{"addRoute", "/api/1/1", "exact", "service-2","", "", },
    pathLookupsTest{"addRoute", "/api/2/1", "prefix", "service-3", "", ""},
    pathLookupsTest{"addRoute", "/api/2/", "prefix", "service-4", "", ""},
    pathLookupsTest{"addRoute", "/api/1", "prefix", "service-5", "", ""},
    pathLookupsTest{"addRoute", "/api/2/1/1", "prefix", "service-6", "", ""},
    pathLookupsTest{"routeLookup", "/api/1", "", "", "service-1", ""},
    pathLookupsTest{"routeLookup", "/api/1/2", "", "", "service-5", ""},
    pathLookupsTest{"routeLookup", "/api/3", "", "", "default-service", ""},
    pathLookupsTest{"routeLookup", "/api/2/1/2", "", "", "service-3", ""},
    pathLookupsTest{"routeLookup", "/api/2/", "", "", "service-4", ""},
    pathLookupsTest{"addRoute", "/api/2/", "prefix", "service-7", "", ""},
    pathLookupsTest{"routeLookup", "/api/2/", "", "", "service-7", ""},

}

// Could structure these tests a bit differently to test the RouteAdds  
// separately from RouteLookups.. functionally this works for now
func TestRouteAddsLookups(t *testing.T){

    for _, test := range addLookupTests{
        if test.operation == "addRoute" {
             output := AddRoute(test.arg1, test.arg2, test.arg3)
             if test.expected1 == "" && output != nil {
                 t.Errorf("Output %q not equal to expected %q", output, test.expected1)
             } 
        }
        if test.operation == "routeLookup" {
             output1, output2 := RouteLookup(test.arg1)
             if output1 != test.expected1 {
                 t.Errorf("Output %q not equal to expected %q", output1, test.expected1)
             } 
             if test.expected2 == "" && output2 != nil {
                 t.Errorf("Output %q not equal to expected %q", output2, test.expected2)
             } 
        }
    }
}

