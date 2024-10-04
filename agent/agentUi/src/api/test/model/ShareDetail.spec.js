/**
 * agent/agentGrpc/agent.proto
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: version not set
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 *
 */

(function(root, factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD.
    define(['expect.js', process.cwd()+'/src/index'], factory);
  } else if (typeof module === 'object' && module.exports) {
    // CommonJS-like environments that support module.exports, like Node.
    factory(require('expect.js'), require(process.cwd()+'/src/index'));
  } else {
    // Browser globals (root is window)
    factory(root.expect, root.AgentAgentGrpcAgentProto);
  }
}(this, function(expect, AgentAgentGrpcAgentProto) {
  'use strict';

  var instance;

  beforeEach(function() {
    instance = new AgentAgentGrpcAgentProto.ShareDetail();
  });

  var getProperty = function(object, getter, property) {
    // Use getter method if present; otherwise, get the property directly.
    if (typeof object[getter] === 'function')
      return object[getter]();
    else
      return object[property];
  }

  var setProperty = function(object, setter, property, value) {
    // Use setter method if present; otherwise, set the property directly.
    if (typeof object[setter] === 'function')
      object[setter](value);
    else
      object[property] = value;
  }

  describe('ShareDetail', function() {
    it('should create an instance of ShareDetail', function() {
      // uncomment below and update the code to test ShareDetail
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be.a(AgentAgentGrpcAgentProto.ShareDetail);
    });

    it('should have the property token (base name: "token")', function() {
      // uncomment below and update the code to test the property token
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property shareMode (base name: "shareMode")', function() {
      // uncomment below and update the code to test the property shareMode
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property backendMode (base name: "backendMode")', function() {
      // uncomment below and update the code to test the property backendMode
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property reserved (base name: "reserved")', function() {
      // uncomment below and update the code to test the property reserved
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property frontendEndpoint (base name: "frontendEndpoint")', function() {
      // uncomment below and update the code to test the property frontendEndpoint
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property backendEndpoint (base name: "backendEndpoint")', function() {
      // uncomment below and update the code to test the property backendEndpoint
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property closed (base name: "closed")', function() {
      // uncomment below and update the code to test the property closed
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

    it('should have the property status (base name: "status")', function() {
      // uncomment below and update the code to test the property status
      //var instance = new AgentAgentGrpcAgentProto.ShareDetail();
      //expect(instance).to.be();
    });

  });

}));