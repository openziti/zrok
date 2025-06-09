package util

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/sirupsen/logrus"
	"os"
	"os/user"
)

// GetHostDetails returns the hostname, detailed host information, and username.
// The detailed host information includes OS, platform, kernel details, and username.
func GetHostDetails() (hostname string, hostDetail string, username string, err error) {
	info, err := host.Info()
	if err != nil {
		return "", "", "", err
	}
	hostname = info.Hostname
	
	userObj, err := user.Current()
	if err == nil && userObj.Username != "" {
		username = userObj.Username
	} else {
		username = os.Getenv("USER")
		if username == "" {
			euid := os.Geteuid()
			username = fmt.Sprintf("user-%d", euid)
			logrus.Warnf("unable to determine the current user, using effective UID: %v", euid)
		}
	}

	hostDetail = fmt.Sprintf("%v; %v; %v; %v; %v; %v; %v",
		info.Hostname, info.OS, info.Platform, info.PlatformFamily,
		info.PlatformVersion, info.KernelVersion, info.KernelArch)
	
	return hostname, hostDetail, username, nil
}

// FormatHostDetailsWithUser combines the username and host details into a single string.
// It also returns a formatted description string if the input description is the default "<user>@<hostname>".
func FormatHostDetailsWithUser(username, hostname, hostDetail, description string) (formattedHostDetail, formattedDescription string) {
	formattedHostDetail = fmt.Sprintf("%v; %v", username, hostDetail)
	
	if description == "<user>@<hostname>" {
		formattedDescription = fmt.Sprintf("%v@%v", username, hostname)
	} else {
		formattedDescription = description
	}
	
	return formattedHostDetail, formattedDescription
} 