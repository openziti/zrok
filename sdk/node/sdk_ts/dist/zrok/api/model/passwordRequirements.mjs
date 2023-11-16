// src/zrok/api/model/passwordRequirements.ts
var _PasswordRequirements = class _PasswordRequirements {
  static getAttributeTypeMap() {
    return _PasswordRequirements.attributeTypeMap;
  }
};
_PasswordRequirements.discriminator = void 0;
_PasswordRequirements.attributeTypeMap = [
  {
    "name": "length",
    "baseName": "length",
    "type": "number"
  },
  {
    "name": "requireCapital",
    "baseName": "requireCapital",
    "type": "boolean"
  },
  {
    "name": "requireNumeric",
    "baseName": "requireNumeric",
    "type": "boolean"
  },
  {
    "name": "requireSpecial",
    "baseName": "requireSpecial",
    "type": "boolean"
  },
  {
    "name": "validSpecialCharacters",
    "baseName": "validSpecialCharacters",
    "type": "string"
  }
];
var PasswordRequirements = _PasswordRequirements;
export {
  PasswordRequirements
};
//# sourceMappingURL=passwordRequirements.mjs.map