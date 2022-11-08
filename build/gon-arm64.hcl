source = ["dist/zrok-arm64_darwin_arm64/zrok"]
bundle_id = "io.zrok.zrok"

apple_id {
	username = "@env:AC_USERNAME"
	password = "@env:AC_PASSWORD"
}

sign {
	application_identity = "Apple Distribution: NetFoundry Inc"
}
