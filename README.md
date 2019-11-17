##################################

README

##################################

A program that can generate music from specified ranges of prometheus metrics. 

GUI:

https://github.com/inkyblackness/imgui-go

Dependencies

##################################

https://github.com/rakyll/portmidi
https://github.com/inkyblackness/imgui-go
https://github.com/go-gl/glfw
https://github.com/prometheus/client_golang

apt-get install libportmidi-dev

go get ./...

Useful posts on threading: 

https://pragmacoders.com/blog/multithreading-in-go-a-tutorial
https://medium.com/@matryer/stopping-goroutines-golang-1bf28799c1cb

Structs: 

https://golangbot.com/structs-instead-of-classes/




Linting/Debug:


      go get -u -v github.com/nsf/gocode
      # OR mdempsky/gocode for better performance
      go get -u -v github.com/mdempsky/gocode
      
      go get -u -v github.com/golang/lint/golint
      go get -u -v golang.org/x/tools/cmd/guru
      go get -u -v golang.org/x/tools/cmd/goimports
      go get -u -v golang.org/x/tools/cmd/gorename

  - **Step 2**: Search and install "Golang Tools Integration" from package control.
  - **Step 3(optional)**: Configure the Settings for `golang` and your project following the `golang.sublime-settings` and `ExampleProject.sublime-project`. Typically, the full features of 'guru' need use the configuration of the project.
  
  ### Tips
  - If you want to trigger auto-completion after ".", you can add below into Settings - Syntax specific - User (a.k.a. User/Go.sublime-settings):
  
      ```json
      {
          "auto_complete_triggers": [{"selector": "source.go - string - comment - constant.numeric", "characters": "."}]
      }
      ```
  
  - If you want to ignore auto-completion when in comments, constant strings, and numbers, you can add below into Settings - Syntax specific - User (a.k.a. User/Go.sublime-settings):
  
      ```json
      {
          "auto_complete_selector": "meta.tag - punctuation.definition.tag.begin, source - comment - string - constant.numeric"
      }
      ```
  
