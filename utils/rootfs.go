package utils

import (
	"fmt"
	"path/filepath"
)

const (
	ImageRoot       = "/var/lib/mydocker/image/"
	OverlayRoot     = "/var/lib/mydocker/overlay2/"
	lowerDirFormat  = OverlayRoot + "%s/lower"
	upperDirFormat  = OverlayRoot + "%s/upper"
	workDirFormat   = OverlayRoot + "%s/work"
	mergedDirFormat = OverlayRoot + "%s/merged"
	overlayFSFormat = "lowerdir=%s,upperdir=%s,workdir=%s"
)

func GetImage(imageName string) string {
	return filepath.Join(ImageRoot, fmt.Sprintf("%s.tar", imageName))
}

func GetRoot(containerID string) string {
	return OverlayRoot + containerID
}

func GetLower(containerID string) string {
	return fmt.Sprintf(lowerDirFormat, containerID)
}

func GetUpper(containerID string) string {
	return fmt.Sprintf(upperDirFormat, containerID)
}

func GetWork(containerID string) string {
	return fmt.Sprintf(workDirFormat, containerID)
}

func GetMerged(containerID string) string {
	return fmt.Sprintf(mergedDirFormat, containerID)
}

func GetOverlayFSDir(containerID string) string {
	return fmt.Sprintf(overlayFSFormat, GetLower(containerID), GetUpper(containerID), GetWork(containerID))
}

func CatOverlayFSDir(lower, upper, work string) string {
	return fmt.Sprintf(overlayFSFormat, lower, upper, work)
}
