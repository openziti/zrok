/*
 * agent/agentGrpc/agent.proto
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * OpenAPI spec version: version not set
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 *
 * Swagger Codegen version: 3.0.51
 *
 * Do not edit the class manually.
 *
 */
(function(root, factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD.
    define(['expect.js', '../../src/index'], factory);
  } else if (typeof module === 'object' && module.exports) {
    // CommonJS-like environments that support module.exports, like Node.
    factory(require('expect.js'), require('../../src/index'));
  } else {
    // Browser globals (root is window)
    factory(root.expect, root.AgentagentGrpcagentproto);
  }
}(this, function(expect, AgentagentGrpcagentproto) {
  'use strict';

  var instance;

  describe('(package)', function() {
    describe('AccessDetail', function() {
      beforeEach(function() {
        instance = new AgentagentGrpcagentproto.AccessDetail();
      });

      it('should create an instance of AccessDetail', function() {
        // TODO: update the code to test AccessDetail
        expect(instance).to.be.a(AgentagentGrpcagentproto.AccessDetail);
      });

      it('should have the property frontendToken (base name: "frontendToken")', function() {
        // TODO: update the code to test the property frontendToken
        expect(instance).to.have.property('frontendToken');
        // expect(instance.frontendToken).to.be(expectedValueLiteral);
      });

      it('should have the property token (base name: "token")', function() {
        // TODO: update the code to test the property token
        expect(instance).to.have.property('token');
        // expect(instance.token).to.be(expectedValueLiteral);
      });

      it('should have the property bindAddress (base name: "bindAddress")', function() {
        // TODO: update the code to test the property bindAddress
        expect(instance).to.have.property('bindAddress');
        // expect(instance.bindAddress).to.be(expectedValueLiteral);
      });

      it('should have the property responseHeaders (base name: "responseHeaders")', function() {
        // TODO: update the code to test the property responseHeaders
        expect(instance).to.have.property('responseHeaders');
        // expect(instance.responseHeaders).to.be(expectedValueLiteral);
      });

    });
  });

}));