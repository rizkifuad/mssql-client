## MSSQL Web Client (Go)

Build the program
```
go build
```

Run

```
ENV=DEV PORT=1515 ./mssql-client
```

Then go to https://localhost:1515

`ENV`  : option `DEV` will update templates without restarting the services
`PORT` : https port, default `4443`

## Packages

* [denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)
* [gorilla/securecookie](https://github.com/gorilla/securecookie)
* [gorilla/mux](https://github.com/gorilla/mux)
* [gorilla/schema](https://github.com/gorilla/schema)
* [jinzhu/gorm](https://github.com/jinzhu/gorm)

## Other

go-vim image taken from [https://github.com/fatih/vim-go/blob/master/assets/vim-go.png](). 
also check out his awesome [vim-go](https://github.com/fatih/vim-go) plugin
