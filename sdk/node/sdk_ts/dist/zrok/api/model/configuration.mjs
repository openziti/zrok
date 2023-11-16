// src/zrok/api/model/configuration.ts
var _Configuration = class _Configuration {
  static getAttributeTypeMap() {
    return _Configuration.attributeTypeMap;
  }
};
_Configuration.discriminator = void 0;
_Configuration.attributeTypeMap = [
  {
    "name": "version",
    "baseName": "version",
    "type": "string"
  },
  {
    "name": "touLink",
    "baseName": "touLink",
    "type": "string"
  },
  {
    "name": "invitesOpen",
    "baseName": "invitesOpen",
    "type": "boolean"
  },
  {
    "name": "requiresInviteToken",
    "baseName": "requiresInviteToken",
    "type": "boolean"
  },
  {
    "name": "inviteTokenContact",
    "baseName": "inviteTokenContact",
    "type": "string"
  },
  {
    "name": "passwordRequirements",
    "baseName": "passwordRequirements",
    "type": "PasswordRequirements"
  }
];
var Configuration = _Configuration;
export {
  Configuration
};
//# sourceMappingURL=configuration.mjs.map