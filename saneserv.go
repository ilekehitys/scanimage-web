package main

// to compile and scp to pi
// env GOARCH=arm GOOS=linux go build saneserv.go && scp saneserv pi11:

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var port = 8080
const html = `
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: monospace;
        }

        select {
            font-family: monospace;
        }
    </style>
</head>
<body onload="readCookies()">
<script>
    function readCookies() {
        var res = getCookie('resolution');
        if (res) {
            console.log("Resolution: " + res)
            document.getElementById('reso').value = res
        }
        var format = getCookie('format');
        if (format) {
            console.log("Format: " + format)
            document.getElementById('format').value = format;
        }
    }

    function getCookie(name) {
        var nameEQ = name + "=";
        var ca = document.cookie.split(';');
        for (var i = 0; i < ca.length; i++) {
            var c = ca[i];
            while (c.charAt(0) == ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
        }
        return null;
    }

    function preview() {
        var theForm = document.forms.theForm;
        var formData = new FormData(theForm);
        var name = formData.get('resolution');
        console.log(name);
        var img = document.createElement("img");
        var preview = document.getElementById('preview');
        preview.innerHTML = "Please wait for preview...<p>".fontcolor('red')
        fetch("preview", {
            method: 'POST',
            body: formData
        }).then(response => response.blob())
            .then(myBlob => {
                img.src = URL.createObjectURL(myBlob);
                while (preview.firstChild) preview.removeChild(preview.firstChild);
                preview.appendChild(img);
            });
    }
</script>
<h1>Scanimage web frontend</h1>
<form action="/" method="post" id="theForm">
    <label for="format">Format:</label>
    <select id="format" name="format">
        <option value="jpeg">jpeg</option>
        <option value="pnm">pnm</option>
        <option value="png">png</option>
        <option value="tiff">tiff</option>
    </select>
    <label for="reso">Resolution:</label>
    <select id="reso" name="resolution">
        <option value="50">50</option>
        <option value="100">100</option>
        <option value="150">150</option>
        <option value="200">200</option>
        <option value="300">300</option>
        <option value="600">600</option>
    </select>
    <p>
        <button type="button" onclick="preview()">Preview (low quality)</Button>
        <input type="submit" value="Submit">
</form>
<div id="preview"></div>
<a href="/err">Check errors</a>
</body>
</html>
`
type FormData struct {
	format, resolution string
}
func ParseFormData (fd *FormData, r *http.Request) error {
	fd.format = r.FormValue("format")
	fd.resolution = r.FormValue("resolution")
	if fd.format != "jpeg" && fd.format != "pnm" && fd.format != "png" && fd.format != "tiff" {
		return errors.New("Invalid format: "+fd.format)
	}
	res, err := strconv.Atoi(fd.resolution)
	if err != nil || (res <= 0 || res > 600) {
		return errors.New("Invalid resolution: "+fd.resolution)
	}
	return nil
}
func scanhandler(w http.ResponseWriter, r *http.Request) {
	dt := time.Now()
	var fd *FormData = &FormData{}
	err := ParseFormData(fd, r)
	if err != nil {
		fmt.Fprintf(w, html)
		return
	}

	cookie1 := &http.Cookie{Name: "format", Value: fd.format, HttpOnly: false, Expires: time.Now().Add(3650 * 24 * time.Hour)}
	cookie2 := &http.Cookie{Name: "resolution", Value: fd.resolution, HttpOnly: false, Expires: time.Now().Add(3650 * 24 * time.Hour)}
	http.SetCookie(w, cookie1)
	http.SetCookie(w, cookie2)
	w.Header().Set("Content-Disposition", "attachment; filename="+dt.Format("01-02-2006_15-04-05")+"."+fd.format)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	cmd := exec.Command("/bin/bash", "-c", "sudo scanimage --format "+fd.format+" --resolution "+fd.resolution)
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Stderr = os.Stderr
	// io.Copy sometimes crashes
	//go io.Copy(w, pipeReader)
	go writeCmdOutput(w, pipeReader)
	cmd.Run()
	pipeWriter.Close()
	fmt.Println(cmd.Stderr)
}

func main() {
	http.HandleFunc("/", scanhandler)
	http.HandleFunc("/err", errorhandler)
	http.HandleFunc("/preview", previewhandler)
	fmt.Println ("Starting listening saneserv at port "+strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func previewhandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		fmt.Println ("Error",err)
		return
	}
	var fd = &FormData{}
	err = ParseFormData(fd, r)
	if err != nil {
		fmt.Println("Invalid form data", err)
		return
	}
	cmd := exec.Command("/bin/bash", "-c", "sudo scanimage --format jpeg --resolution 50")
	w.Header().Set("Content-Type", "image/png")
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Stderr = os.Stderr
	// io.Copy sometimes crashes
	//go io.Copy(w, pipeReader)
	go writeCmdOutput(w, pipeReader)
	cmd.Run()
	pipeWriter.Close()
}
func writeCmdOutput(res http.ResponseWriter, pipeReader *io.PipeReader) {
	buffer := make([]byte, 8)
	for {
		n, err := pipeReader.Read(buffer)
		if err != nil {
			pipeReader.Close()
			break
		}
		data := buffer[0:n]
		res.Write(data)
		if _, ok := res.(http.Flusher); ok {
			//causes program to crash
			//f.Flush()
		}
		// reset buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

func errorhandler(writer http.ResponseWriter, request *http.Request) {
	cmd := exec.Command("/bin/bash", "-c", "sudo systemctl status saneserv 2>&1")
	out, _ := cmd.Output()
	writer.Write(out)
}
