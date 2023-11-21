"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
const { Command } = require("commander");
const zrok = require("zrok");
const ziti = require('@openziti/ziti-sdk-nodejs');
const express = require('express');
const program = new Command();
program
    .command('copyto')
    .version("1.0.0")
    .description("command to host content to be pastedfrom'd")
    .action(() => __awaiter(void 0, void 0, void 0, function* () {
    let root = zrok.Load();
    //await ziti.init( root.env.ZitiIdentity ).catch(( err: Error ) => { console.error(err); return process.exit(1) });
    ziti.setLogLevel(10);
    let shr = yield zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PUBLIC_SHARE_MODE, "pastebin", ["public"]));
    console.log("setting up app");
    let service = "ns5ix2brb61f";
    console.log("attempting to bind to service: " + service);
    let app = ziti.express(express, service);
    console.log("after setting up app");
    app.get('/', function (_, res) {
        res.write("Test");
    });
    console.log("after setting up get");
    //app.listen(undefined, () => {
    //  console.log(`Example app listening!`)
    //})
    console.log("after listen");
    zrok.DeleteShare(root, shr);
}));
program
    .command('pastefrom <shrToken>')
    .version("1.0.0")
    .description("command to paste content from coptyo")
    .action((shrToken) => {
    console.log('pastefrom command called', shrToken);
});
program.parse(process.argv);
const options = program.opts();
//# sourceMappingURL=index.js.map