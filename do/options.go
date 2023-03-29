package do

func getRegionFromOption(region string) string {
	switch region {
	case "usw":
		return "sfo1"
	case "use":
		return "nyc1"
	case "eu":
		return "fra1"
	case "ap":
		return "sgp1"
	case "nz":
		return "syd1"
	default:
		return ""
	}
}

func getSizeFromOption(size string) string {
	switch size {
	case "xs":
		return "s-1vcpu-512mb-10gb"
	case "s":
		return "s-1vcpu-1gb"
	case "m":
		return "s-2vcpu-4gb"
	case "l":
		return "s-4vcpu-8gb"
	case "xl":
		return "s-8vcpu-16gb"
	default:
		return ""
	}
}
