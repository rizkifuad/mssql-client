function readCookie(name) {
  var nameEQ = name + "=";
  var ca = document.cookie.split(';');
  for(var i=0;i < ca.length;i++) {
    var c = ca[i];
    while (c.charAt(0)==' ') c = c.substring(1,c.length);
    if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
  }
  return null;
}

var getFakeStore = function() {
  var store = {}
  if (localStorage.getItem("hash") === readCookie("hash")) {
    if (localStorage.getItem("fakeStore")) {
      store = JSON.parse(localStorage.getItem("fakeStore"))
    }

    return store
  }

  localStorage.setItem("hash", readCookie("hash"))
  return null
}

var setFakeStore = function(data) {
  var store = {}
  for (i in data) {
    if (typeof data[i] !== "function") {
      store[i] = data[i]
    }
  }
  store = JSON.stringify(store)
  localStorage.setItem("fakeStore", store)
}

var fakeStore = getFakeStore()

var connectionTemplate = {
  name: "New Connection",
  server: "",
  port: 1433,
  user: "",
  password: "",
  query: "",
  cols: [],
  rows: [],
  error: false,
  message: "",
  fetching: false,
  databases: [],
  database: "",
  status: "init" //init, connected, disconnect, error
}

var initTemplate = {
  id: 0,
  name: "",
  server: "",
  port: 1433,
  user: "",
  password: "",
  database: "",
  fetching: false
}

new Vue({
  el:"#root",
  components: {
    Multiselect: window.VueMultiselect.default
  },
  delimiters: ['${', '}'],
  data: {
    //active connection data
    activeConnections: [],
    active: null,
    currentTab: 0,

    //init data
    savedConnections: [],
    init: {},

    //search connection
    search: ""
  },

  updated() {
    setFakeStore(this._data)
  },

  created() {
    this.init = Object.assign({}, initTemplate)
    if (fakeStore) {
      for(i in fakeStore) {
        this[i] = fakeStore[i]
      }

      this.active = this.activeConnections[this.currentTab]
    } else {
      this.activeConnections[this.currentTab] = Object.assign({}, connectionTemplate)
      this.active = this.activeConnections[this.currentTab]

    }

    var self = this
    axios.get("/api/get_connections")
      .then( (conns) =>  {
        self.savedConnections = conns.data.data
      })
  },


  methods: {
    reset() {
      this.active.cols = []
      this.active.rows = []

      this.active.error = false
      this.active.message = ""
    },
    onQuery() {
      this.fetching = true
      this.reset()

      var run = axios({
        method:"POST",
        url: "/api/query",
        data: parseUrlEncoded({
          query: this.active.query,
          id: this.currentTab
        }),
        headers: {"content-type": "application/x-www-form-urlencoded"}
      })

      run.then( (resp) => {
        this.active.fetching = false
        if (resp.status) {
          var response = resp.data
          this.active.cols = response.data.cols
          this.active.rows = response.data.rows ? response.data.rows : []

          var rowsCount = this.active.rows ? this.active.rows.length : 0
          this.active.error = false
          this.active.message = `No errors; ${rowsCount} rows affected; taking ${response.data.elapsed}`
        }
      })

      run.catch( (err, data) => {
        this.active.fetching = false
        this.active.error = true
        this.active.message = `Error executing query: ${err.response.data.message}`
      })
    },

    changeDatabase() {
      var changeDB = axios({
        method: "POST",
        url: "/api/change_database",
        headers: {"content-type": "application/x-www-form-urlencoded"},
        data: parseUrlEncoded({
          id: this.currentTab,
          database: this.active.database
        })
      })

      changeDB.then( (resp) => {
        this.active.fetching = false
      })
    },

    exportCSV() {
      if (this.active.rows.length === 0) {
        this.showModal = true
        return
      }
      var data = [[this.active.cols]]
      for (i in this.active.rows) {
        data.push([this.active.rows[i]])
      }
      var csvContent = "data:text/csv;charset=utf-8,";
      data.forEach(function(infoArray, index){
        dataString = infoArray.join(",");
        csvContent += index < data.length ? dataString+ "\n" : dataString;
      }); 

      var encodedUri = encodeURI(csvContent);
      var link = document.createElement("a");
      link.setAttribute("href", encodedUri);
      var now = new Date()
      link.setAttribute("download", this.active.name + "-" + this.active.database + "-" + now.getFullYear() + "-" + addNol(now.getMonth()) + "-" + addNol(now.getDate()) + "_" + addNol(now.getHours()) + "." +
        addNol(now.getMinutes()) + "." + addNol(now.getSeconds()) + ".csv");
      document.body.appendChild(link);

      link.click();

    },


    closeModal() {
      this.showModal = false
    },

    selectConnection(id) {
      this.init = this.savedConnections[id]
    },

    saveConnection() {
    },

    openConnection() {
      var updated = ['server', 'port', 'user', 'password', 'database', 'name']
      delete this.init.databases

      var run = axios({
        method:"POST",
        url: "/api/add_connection",
        data: parseUrlEncoded(this.init),
        headers: {"content-type": "application/x-www-form-urlencoded"}
      })

      run.then( (resp) => {
        for (i in updated) {
          this.activeConnections[this.currentTab][updated[i]] = this.init[updated[i]]
        }

        this.activeConnections[this.currentTab].status = "connect"
        this.active = this.activeConnections[this.currentTab]
        this.active.databases = resp.data.data.databases

        for(i in this.active.databases) {
          var dbname = this.active.databases[i].name
          if (dbname.toLowerCase() === this.active.database.toLowerCase()) {
            this.active.database = dbname
            return
          }
        }

      })

    },

    closeConnection(id) {
      this.activeConnections.splice(id, 1)
      var del = axios({
        method: "DELETE",
        url: "/api/disconnect",
        headers: {"content-type": "application/x-www-form-urlencoded"},
        data: parseUrlEncoded({
          id: id,
        })
      })

      del.then( (resp) => {
        if (this.currentTab === id) {
          if (this.activeConnections.length === 0) {
            this.currentTab = 0
            this.activeConnections[this.currentTab] = Object.assign({}, connectionTemplate)
          } else {
            if (this.activeConnections[this.currentTab-1]) {
              this.currentTab = this.currentTab - 1
            } else if (this.activeConnections[this.currentTab+1]) {
              this.currentTab = this.currentTab + 1
            }
          }
        } else {
          if (this.currentTab > id) {
            this.currentTab = this.currentTab - 1
          }
        }
        this.active = this.activeConnections[this.currentTab]
      })

    },

    addConnection() {
      this.resetInit()
      this.currentTab = this.activeConnections.length
      this.activeConnections[this.currentTab] = Object.assign({}, connectionTemplate)
      this.active = this.activeConnections[this.currentTab]
    },

    changeTab(id) {
      this.currentTab = id
      this.active = this.activeConnections[this.currentTab]
    },

    resetInit() {
      this.init = Object.assign({}, initTemplate)
    }

  },

  computed: {
    filteredSavedConnections: function() {
      if (!this.search) {
        return this.savedConnections
      }

      var search = this.search.toLowerCase().trim()
      var regex = new RegExp("^.*"+search+".*$","g");


      var conns = this.savedConnections.filter( (conn, id) => {
        return regex.test(conn.name.toLowerCase())
      })

      return conns
    },

    parsedInitDatabases() {
      var dbs = []
      for (i in this.init.databases) {
        dbs.push(this.init.databases[i].name)
      }
      return dbs
    },

    parsedActiveDatabases() {
      var dbs = []
      for (i in this.active.databases) {
        dbs.push(this.active.databases[i].name)
      }
      return dbs
    }
  }
})

var addNol = function(data) {
  data = data + ""
  if (data.length === 1) {
    data = "0" + data
  }

  return data
}

var parseUrlEncoded = function(data) {
  var params = new URLSearchParams()
  for (key in data) {
    params.append(key, data[key])
  }
  return params
}
