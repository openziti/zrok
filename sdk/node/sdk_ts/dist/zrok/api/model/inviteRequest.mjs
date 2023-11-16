// src/zrok/api/model/inviteRequest.ts
var _InviteRequest = class _InviteRequest {
  static getAttributeTypeMap() {
    return _InviteRequest.attributeTypeMap;
  }
};
_InviteRequest.discriminator = void 0;
_InviteRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var InviteRequest = _InviteRequest;
export {
  InviteRequest
};
//# sourceMappingURL=inviteRequest.mjs.map