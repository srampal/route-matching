/* Code exercise program.  Define a library that allows creation of "Routes" with paths and associated service strings and also allows subsequent retrieval of best matchng routes. Each route can be of type "exact" or "prefix" corresponding to exact match lookup and longest prefix match based lookup. The service corresponding to the best match should be returned always and if there is no match, a default service called "default-service" should be returned. 

Exact match routes get priority over prefix match routes.  Also, AddRoute() can be used to modify the "destination Service" for an existing entry, however AddRoute for the same path, once as a exact match route and once as a prefix match route, should not be treated as a modification of the first route and should allow both routes to exist with the exact match route having priority over the prefix match version

Design summary:
A combination of a map/ hash table (for the exact match routes) and a sorted list (slice) (for the prefix match routes) is used.  Entries are first looked up in the exact match table. If found in this table, we are done (usually with an O(1) lookup).  If not found in the exact match table, the path is then looked up in the prefix match slice which is sorted by decreasing prefix length (longest prefix first)  and looked up with a sequential traversal.  Once a match is found in the prefix list, a dynamically generated exact match entry is also created in the exact match hash table so that future lookups for this path can again in the fast path.  Anytime a prefix entry is created/ deleted modified, the corresponding dynamically generated exact match entries are deleted/ flushed since they may now have stale routing information. 

The current design should be well performant specially due to the creation of the dynamic "cache" style entries which will convert must lookups to be an O(1) operation for any mix of exact match and prefix match entries. If additional performance is desired, it could be achieved via alternate data structures including radix/ Patricia trees and arrays of slices for the prex table. A simple initial enhancement would be to change the slice to a linked list so that insertions and deletions are faster and there is no requirement for large contiguous memory as is the case with arrays or slices.


Current caveats include: 
1) Trailing "/" identified a separate path (in practice "xyz" and "xyz/" would usually be considered the same path)
2) Path is considered case sensitive  
3) Input validation checks are not applied as they would in practice
4) Route adds and modifies are supported but not deletions, this can easily be added
5) Although the performance should be reasonable given the use of dynamically created exact match entries, even further 
performance optimizations are possible (including changing the prefix match list to either an array of lists or  
trie structure that can be used to perform longest prefix string matching without needing to traverse a single list). 
Optimizations for table memory scale, and usage/ hit rate based caching can also be added to address deployment environment 
constraints.

e.g. AddRoute("/api/1", "exact", "service-1")
     AddRoute("/api/1/1", "exact", "service-2")
     AddRoute("/api/2/1", "prefix", "service-3")
     AddRoute("/api/2/", "prefix", "service-4")
     AddRoute("/api/1", "prefix", "service-5")     // co-exists with exact match route of same path
     AddRoute("/api/2/1/1", "prefix", "service-6")

     RouteLookup("/api/1")     should return "service-1"
     RouteLookup("/api/1/2")   should return "service-5"
     RouteLookup("/api/3")     should return "default-service"
     RouteLookup("/api/2/1/2"  should return "service-3"
     RouteLookup("/api/2/"     should return "service-4"
     AddRoute("/api/2/", "prefix", "service-7"     // modifies the existing route from service-4 to service-7
     RouteLookup("/api/2/"     should return "service-7"

*/
package main

   import "fmt"
   import "sort"

// Structure defining each route whether staticaly created by external users or dynamically
// generated internally to cache results from prefix table for faster lookups
type route struct {
    path               string              // the path of the route
    matchType          string              // "exact" or "prefix"
    destination        string              // the target/ destination service 
    dynamicEntry       bool                // true => internally dynamically created entry
                                           // false => user provided input "static" entry
}

// Table (slice) of pointers to prefix match routes
var prefixRoutesTable  []route             // Note: we use the term Table instead of list since in future 
                                           // it could be implemented  with another data structure 

// Table (hash table/ map) for exact match routes
var exactMatchTable  map[string]route     // Note: we use the term Table instead of map or hash etc since
                                           // in future it could be implemented with another data structure

//Table of paths/ keys  of  dynamically created exact match routes   
var dynamicEntries  []string              // This is used to flush dynamically created entries that may be stale


//Initialize the tables (maps, slices etc) needed by the rest of the functions in the package
func init() {

     prefixRoutesTable = make([]route, 10)
     exactMatchTable = make(map[string]route, 10)
     dynamicEntries  = make([]string, 10)
}

func main() {

     if err := AddRoute("/api/1", "exact", "service-1"); err != nil {
         fmt.Println(err)
     }

     if err := AddRoute("/api/1/1", "exact", "service-2"); err != nil {
         fmt.Println(err)
     }

     if err := AddRoute("/api/2/1", "prefix", "service-3"); err != nil {
         fmt.Println(err)
     }

     if err := AddRoute("/api/2/", "prefix", "service-4"); err != nil {
         fmt.Println(err)
     }

     if err := AddRoute("/api/1", "prefix", "service-5"); err != nil {
         fmt.Println(err)
     }

     if err := AddRoute("/api/2/1/1", "prefix", "service-6"); err != nil {
         fmt.Println(err)
     }

     if result, err := RouteLookup("/api/1"); err != nil {
         fmt.Println("Error looking up /api/1 ", err)
     } else {
         fmt.Println("Lookup for /api/1 is ", result)
     } 

     if result, err := RouteLookup("/api/1/2"); err != nil {
         fmt.Println("Error looking up /api/2 ", err)
     } else {
         fmt.Println("Lookup for /api/1/2 is ", result)
     } 

     if result, err := RouteLookup("/api/3"); err != nil {
         fmt.Println("Error looking up /api/3 ", err)
     } else {
         fmt.Println("Lookup for /api/3 is ", result)
     } 

     if result, err := RouteLookup("/api/2/1/2"); err != nil {
         fmt.Println("Error looking up /api/2/1/2 ", err)
     } else {
         fmt.Println("Lookup for /api/2/1/2 is ", result)
     } 

     if result, err := RouteLookup("/api/2/"); err != nil {
         fmt.Println("Error looking up /api/2/ ", err)
     } else {
         fmt.Println("Lookup for /api/2/ is ", result)
     } 

     if err := AddRoute("/api/2/", "prefix", "service-7"); err != nil {
         fmt.Println(err)
     }

     if result, err := RouteLookup("/api/2/"); err != nil {
         fmt.Println("Error looking up /api/2/ ", err)
     } else {
         fmt.Println("Lookup for /api/2/ is ", result)
     } 

     printAllTables(0)

     return
}

func printAllTables(i int) {
    fmt.Printf("%d) \n", i)
    fmt.Println("prefixRoutesTable")
    fmt.Println(prefixRoutesTable)
    fmt.Println("exactMatchTable")
    fmt.Println(exactMatchTable)
    fmt.Println("dynamicEntries")
    fmt.Println(dynamicEntries)
}


// AddRoute is called to add a new route or modify an existing route with behavior as described in the package comments
func AddRoute (path string, matchType string, destination string) error {

// Perform input validation here (skipped for now)

   rte := route {
    	path: path, 
    	matchType: matchType,
    	destination: destination,
    	dynamicEntry: false,
   }

// If this is an exact route, insert into exactMatchTable (modify also automatically handled)
// Note: any prior entry is freed and garbage collected, (could also add explicit freeing 
// as an optimization to not depend on the garbage collector)

   if matchType == "exact" {
      exactMatchTable[path] = rte   
      return nil
   }

   // So this is a prefix route, first check if it exists already
   // If it does exist and the destination has changed, update it, flush dynamic entries 

   for i, r := range prefixRoutesTable {
       if r.path == path {
           prefixRoutesTable[i].destination = destination
           defer flushDynamicRoutes()  
           return nil 
       }
   }

   // Not found in the prefix table either, so add it as a new entry and re-sort the slice 

   prefixRoutesTable = append(prefixRoutesTable, rte)

   sort.Slice(prefixRoutesTable, func(i, j int) bool { return len(prefixRoutesTable[i].path) > 
                                                       len(prefixRoutesTable[j].path) } ) 

   return nil
}

// Go through all entries (paths) in the dynamicEntries slice and delete the dynamic entries created
// for them in the exactMatchTable
func flushDynamicRoutes() {
   for _, p := range dynamicEntries {
       if p != "" {
           dR, ok := exactMatchTable[p]
           if ok && dR.dynamicEntry {
               delete(exactMatchTable, p)
           } 
       }
   }

   // Finally clean up the dynamicEntries slice as well
   dynamicEntries = nil

   return
}

// RouteLookup is called to return the destination service associated with the best match route for the input path provided
func RouteLookup (path string) (string, error) {

// Lookup in the exact match table, if found, we are done 
      
      rte, ok := exactMatchTable[path] 

      if ok {
            fmt.Println("Lookup result -> ", rte.destination)
            return rte.destination, nil
      }

// else searchInPrefixTable (sequential walk and prefix match check)

   for _, r := range prefixRoutesTable {
       // If the input path's length is equal or greater than the path for the existing route and 
       // matches up to the length of the route's path, we have a match
       
       if (len(path) >= len(r.path)) && len(r.path) != 0 && (r.path == path[0:len(r.path)]) {
           fmt.Println("Lookup result -> ", r.destination)
           createDynamicRoute(path, r.destination)
           return r.destination, nil
       } 
   }

   // If still not found, the destination defaults to the default service
   fmt.Println("Lookup result -> default-service")
   createDynamicRoute(path, string("default-service"))
   return "default-service", nil

}

func createDynamicRoute(path string, destination string) {

  // First create an entry in the dynamicEntries slice  
  dynamicEntries = append(dynamicEntries, path)

  // Then create an entry in the exact MatchTable for this path and mark it as a dynamic entry
  rte := route {
    path: path,
    matchType: "exact",
    destination: destination,
    dynamicEntry: true,
  }

  exactMatchTable[path] = rte

}


