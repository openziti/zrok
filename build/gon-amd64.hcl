source = ["dist/zrok-amd64_darwin_amd64_v1/zrok"]
bundle_id = "io.zrok.zrok"

apple_id {
	username = "@env:AC_USERNAME"
	password = "@env:AC_PASSWORD"
}

sign {
	application_identity = "Apple Distribution: NetFoundry Inc"
}
