# Scanimage web frontend 

This go program listens for http requests and provides a simple scanimage (sane) frontend to usb scanner. First connect your usb scanner for example to raspberry pi. Then compile the program and copy the binary to the server. Start binary by command line and point your browser to the listening port.  

You need to install and configure sane and scanimage first. Check that scanimage works from command line. If you are able to run the scanimage command without root, remove "sudo" from command line.  

My scanner, Samsung SCX-3200, needs to have a little idle moment between scans, otherwise an empty image is downloaded. Also most of scanimage command line switches seem to be unusable or not supported so I left them out. However, basic scanning seem to work.  

Good information about sane and scanimage is available here:   

https://wiki.archlinux.org/index.php/SANE  

Below is a screenshot. Cookies are stored to remember last format and resolution settings. Tested browsers are Chrome version 84.0.4147.89 and Firefox 78.0.2.   

![Alt text](preview.png?raw=true "Web frontend")
