package inode

import (
	"os/exec"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	r = regexp.MustCompile(`(?s)(\d+) (\d+) (\d+) +(\d+)\%`)
)

func run(args ...string) (string, error) {
	log.Debugf("inode.run(\"%v\")", args)
	baseCmd := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(baseCmd, cmdArgs...)
	out, err := cmd.Output()
	return string(out), err
}

func GetInodesInfo(path string) (int, int, int, int, error) {
	log.Debugf("inode.GetInodesInfo(\"%v\")", path)
	output, err := run("df", "--output=itotal,iused,iavail,ipcent", path)

	if err != nil {
		return 0, 0, 0, 0, err
	}

	s := r.FindStringSubmatch(output)

	log.WithFields(log.Fields{
		"output":                       output,
		"r.FindStringSubmatch(output)": s,
	}).Trace("inode.GetInodesInfo.r.FindStringSubmatch")

	var (
		itotal, iused, iavail, ipcent int
	)

	itotal, err = strconv.Atoi(s[1])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	iused, err = strconv.Atoi(s[2])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	iavail, err = strconv.Atoi(s[3])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	ipcent, err = strconv.Atoi(s[4])
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return itotal, iused, iavail, ipcent, nil
}
