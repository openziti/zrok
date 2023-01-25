source = ["dist/zrok-amd64_darwin_amd64_v1/zrok"]
bundle_id = "io.zrok.zrok"

apple_id {
	password = "@env:AC_PASSWORD"
}

sign {
	application_identity = "Developer ID Application: NetFoundry Inc"
}

zip {
	output_path = "dist/zrok-amd64_darwin_amd64_v1/zrok.zip"
}