source = ["dist/zrok-arm64_darwin_arm64/zrok"]
bundle_id = "io.zrok.zrok"

apple_id {
	password = "@env:AC_PASSWORD"
}

sign {
	application_identity = "Developer ID Application: NetFoundry Inc"
}

zip {
	output_path = "dist/zrok-arm64_darwin_arm64/zrok.zip"
}