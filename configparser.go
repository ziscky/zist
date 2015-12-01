package main

import (
	"encoding/xml"
	"os"
)

type App struct {
	Name xml.Name    `xml:"app"`
	Ep   Endpoint    `xml:"endpoint"`
	Cd   Credentials `xml:"credentials"`
	Rp   RProxy      `xml:"rproxy"`
	Ps   Processes   `xml:"processes"`
}
type Endpoint struct {
	Name     xml.Name `xml:"endpoint"`
	Protocol string   `xml:"protocol"`
	Port     string   `xml:"port"`
	Token    string   `xml:"token"`
}

type Credentials struct {
	Name        xml.Name           `xml:"credentials"`
	MailCred    mailcredentials    `xml:"mail"`
	TwitterCred twittercredentials `xml:"twitter"`
}
type mailcredentials struct {
	Name     xml.Name `xml:"mail"`
	Enabled  bool     `xml:enable,attr`
	Address  string   `xml:"address"`
	Uname    string   `xml:"uname"`
	Password string   `xml:"password"`
}
type twittercredentials struct {
	Name           xml.Name `xml:"twitter"`
	Enable         bool     `xml:"enable,attr"`
	ConsumerKey    string   `xml:"consumerkey"`
	ConsumerSecret string   `xml:"consumersecret"`
	AccessKey      string   `xml:"accesskey"`
	AccessSecret   string   `xml:"accesssecret"`
}

type RProxy struct {
	Name    xml.Name `xml:"rproxy"`
	Proxies []proxy  `xml:"proxy"`
}

type proxy struct {
	Name xml.Name `xml:"proxy"`
	Type string   `xml:"type"`
	From string   `xml:"from"`
	To   string   `xml:"to"`
}
type Processes struct {
	Name xml.Name  `xml:"processes"`
	Pss  []process `xml:"process"`
}
type process struct {
	Name        xml.Name    `xml:"process"`
	Type        string      `xml:"type,attr"`
	Pname       string      `xml:"pName"`
	Path        string      `xml:"pPath"`
	Restart     string      `xml:"restart"`
	Plogs       plogs       `xml:"pLogs"`
	Outputs     outputs     `xml:"out"`
	CrashReport crashreport `xml:"crashreporting"`
}
type plogs struct {
	Name  xml.Name `xml:"pLogs"`
	Plogs []plog   `xml:"log"`
}
type plog struct {
	Name     xml.Name `xml:"log"`
	Path     string   `xml:"path"`
	Interval int      `xml:"interval"`
	Twitter  string   `xml:"twitter"`
	Mail     string   `xml:"mail"`
}
type outputs struct {
	Name   xml.Name      `xml:"out"`
	StdOut stdouputs     `xml:"stdout"`
	StdErr stderroutputs `xml:"stderr"`
	Stats  statoutputs   `xml:"stats"`
}
type stdouputs struct {
	Name     xml.Name `xml:"stdout"`
	Web      string   `xml:"web"`
	Twitter  string   `xml:"twitter"`
	Interval int      `xml:"interval"`
}
type stderroutputs struct {
	Name     xml.Name `xml:"stderr"`
	Web      string   `xml:"web"`
	Twitter  string   `xml:"twitter"`
	Interval int      `xml:"interval"`
}
type statoutputs struct {
	Name xml.Name `xml:"stats"`
	Web  string   `xml:"web"`
}
type crashreport struct {
	Name    xml.Name      `xml:"crashreporting"`
	Enable  bool          `xml:"enable,attr"`
	Twitter twitterreport `xml:"twitter"`
	Mail    mailreport    `xml:"mail"`
}
type twitterreport struct {
	Name    xml.Name `xml:"twitter"`
	Message string   `xml:"message"`
	Hashtag string   `xml:"hashtag"`
}
type mailreport struct {
	Name    xml.Name     `xml:"mail"`
	Rcps    []recepients `xml:"tos"`
	Body    string       `xml:"body"`
	Subject string       `xml:"subject"`
}
type recepients struct {
	Name       xml.Name `xml:"tos"`
	Recepients []string `xml:"to"`
}

func ParseXMLDirectives(path string) (*App, error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}
	defer file.Close()
	app := new(App)
	err = xml.NewDecoder(file).Decode(app)
	if err != nil {
		return nil, err
	}
	return app, nil
}
