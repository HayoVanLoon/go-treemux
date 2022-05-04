# Tree-based Request Multiplexer

*Experimental, may change in backwards-incompatible ways*

TreeMux is an HTTP request multiplexer that routes using a tree structure.

Wildcards ("*") are used to indicate flexible path elements in a resource URL,
which can then be mapped to a single Handler (function).

# Example
With the following route:

```go
t.Handle("/countries/*/cities", handleCities)
```

Paths like these will be handled by `handleCities`:

```text
"/countries/belgium/cities"
"/countries/france/cities"
```

There is no support for elements with partial wildcards (i.e. `/foo*/bar`).

# License

Copyright 2022 Hayo van Loon

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the
License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
