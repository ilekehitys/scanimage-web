# Scanimage web frontend 

This go program listens for http requests and provides a simple scanimage (xane) frontend to usb scanner. First connect your usb scanner for example to raspberry pi. Then compile the program and copy the binary to the server. Start binary by command line and point your browser to the listening port.  

You need to install and configure xane and scanimage first. Check that scanimage works from command line. If you are able to run the scanimage command without root, remove "sudo" from source code.  

My scanner, Samsung SCX-3200, needs to have a little break between scans. Also most of scanimage switches seem to be unusable or not supported so I left them out. However, basic scanning seem to work.  

![Alt text](preview.png?raw=true "Web frontend")
