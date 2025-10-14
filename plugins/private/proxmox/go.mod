module proxmox-plugin

go 1.24.0

require skyclust v0.0.0-00010101000000-000000000000

replace skyclust => ../../../

require github.com/luthermonson/go-proxmox v0.2.3

require (
	github.com/buger/goterm v1.0.4 // indirect
	github.com/diskfs/go-diskfs v1.5.0 // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jinzhu/copier v0.3.4 // indirect
	github.com/magefile/mage v1.14.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)
