package rfile

import (
	"bufio"
	"compress/gzip"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/sftp"
)

const gzipfile = ".gz"

// File structure used to select nmon files to import
type File struct {
	Name     string
	FileType string
	log      *logrus.Logger // The sytem Log
	sftpConn *sftp.Client
	reader   *RemoteFileReader
}

// RemoteFileReader struct for remote files
type RemoteFileReader struct {
	*sftp.File
	*bufio.Reader
}

// New Create a RemoteFile
func New(sftp *sftp.Client, l *logrus.Logger, name string) *File {
	f := &File{Name: name, log: l, sftpConn: sftp}
	f.Init()
	return f
}

// Init Initialice File Reader
func (rf *File) Init() {
	reader, err := rf.GetRemoteReader()
	if err != nil {
		rf.log.Errorf("Error on Remote Reader :%s", err)
		return
	}
	rf.reader = reader
}

// End close RemoteFile
func (rf *File) End() error {
	rf.reader.Close()
	return nil
}

// GetRemoteReader open an nmon file based on file extension and provides a bufio Reader
func (rf *File) GetRemoteReader() (*RemoteFileReader, error) {
	rf.log.Debugf("Open remote file %s", rf.Name)
	file, err := rf.sftpConn.Open(rf.Name)
	if err != nil {
		return nil, err
	}
	rf.log.Debugf("Remote reader %#+v", file)

	if rf.FileType == gzipfile {
		gr, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		reader := bufio.NewReader(gr)
		//return &RemoteFileReader{file, bufio.NewReader(reader)}, nil
		return &RemoteFileReader{file, reader}, nil
	}

	reader := bufio.NewReader(file)
	//return &RemoteFileReader{file, bufio.NewReader(reader)}, nil
	return &RemoteFileReader{file, reader}, nil
}

//Content returns the nmon files content sorted in an slice of string format
func (rf *File) Content() []string {

	var lines []string
	for {
		line, _, err := rf.reader.ReadLine()
		if err != nil {
			break
		}
		lines = append(lines, string(line))
	}

	return lines
}
