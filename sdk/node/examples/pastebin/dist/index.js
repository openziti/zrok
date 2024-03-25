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
Object.defineProperty(exports, "__esModule", { value: true });
const { Command } = require("commander");
const zrok = require("zrok");
const ziti = require('@openziti/ziti-sdk-nodejs');
const express = require('express');
var readlineSync = require('readline-sync');
const program = new Command();
program
    .command('copyto')
    .version("1.0.0")
    .description("command to host content to be pastedfrom'd")
    .action(() => __awaiter(void 0, void 0, void 0, function* () {
    var data = readlineSync.question('Input some text... ');
    console.log("data is: ", data);
    let root = zrok.Load();
    yield zrok.init(root);
    //await ziti.init( root.env.ZitiIdentity ).catch(( err: Error ) => { console.error(err); return process.exit(1) });
    ziti.setLogLevel(0);
    console.log("setting up zrok.CreateShare...");
    let shr = yield zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PUBLIC_SHARE_MODE, "pastebin", ["public"]));
    // console.log("zrok share is: ",shr)
    // console.log("setting up app")
    // let service = "ns5ix2brb61f"
    // console.log("attempting to bind to service: "+ shr.Token)
    console.log("access your pastebin using 'pastefrom ", shr.Token);
    // let listener = await zrok.NewListener(shr.Token, root)
    // console.log("zrok listener is:  ", listener)
    let app = ziti.express(express, shr.Token);
    // console.log("after setting up app")
    app.get('/', function (_, res) {
        // console.log("received a GET request")
        res.write(data);
        res.end();
    });
    // console.log("after setting up get")
    app.listen(undefined, () => {
        // console.log(`Example app listening!`)
    });
    // console.log("after listen")
    // zrok.DeleteShare(root, shr);
}));
program
    .command('pastefrom <shrToken>')
    .version("1.0.0")
    .description("command to paste content from coptyo")
    .action((shrToken) => __awaiter(void 0, void 0, void 0, function* () {
    // ziti.setLogLevel(99)
    let root = zrok.Load();
    yield zrok.init(root).catch((err) => {
        console.log(err);
    });
    let acc = yield zrok.CreateAccess(root, new zrok.AccessRequest(shrToken));
    // console.log("zrok.CreateAccess returned: ", acc)
    // console.log("about to ziti.httpRequest: ", shrToken)
    // setTimeout(function() {
    ziti.httpRequest(shrToken, undefined, 'GET', '/', [], (data) => {
        console.log("in on_req_cb");
        console.log("data is: ", data);
    }, (data) => {
        console.log("in on_resp_cb");
        console.log("data is: ", data);
    }, (data) => {
        console.log("in on_resp_data_cb");
        console.log("data is: ", data);
        if (data.body) {
            console.log('----------- pastefrom is: ', data.body.toString());
            process.exit(0);
        }
    });
    // }, 5000);
    // zrok.dialer(
    //   root, 
    //   shrToken, 
    //   (conn: any) => {
    //     console.log("in connectCallback")
    //     console.log("conn is: ", conn)
    //   },
    //   (dataData: any) => {
    //     console.log("in dataCallback")
    //     console.log(dataData.toString())
    //   }
    // );
}));
program.parse(process.argv);
const options = program.opts();
//# sourceMappingURL=index.js.map