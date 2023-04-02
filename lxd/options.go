package lxd

func getProfileFromSizeOption(size string) string {
	return LXD_PROFILE_PREFIX + size
}
