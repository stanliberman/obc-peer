This a sample asset management chaincode.

Assumptions:
	- Single asset per chaincode
	- No metadata for transactions or asset
	- OBC doesn't pass in the credentials of the requestor, so we pass in the user id as the first argument

	TODO Users not present in the system cannot query the chaincode
	TODO Read permissions
