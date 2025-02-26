package runtime

var CurrentView = "main"
var ErrorMsg = ""
var ErrorFullTrace = ""

var UpdateList = false

var LoginStatus = make(map[string]interface{})
var Data = make(map[string]interface{})
