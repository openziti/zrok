"use client";

import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "@/api/src";
import {Table, TableBody, TableCell, TableColumn, TableHeader, TableRow} from "@nextui-org/table"

export default function Home() {
    const [version, setVersion] = useState("");
    const [status, setStatus] = useState({});
    let api = new AgentApi(new ApiClient("http://localhost:8888"));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
            console.log("error", err);
            console.log("data", data);
            if(mounted) {
                setVersion(data.v);
            }
        });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                console.log(data);
                if(mounted) {
                    setStatus(data);
                }
            })
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    return (
        <div className="grid grid-rows-[20px_1fr_20px] p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
            <h1>Agent: {version}</h1>

            <div>
                <h2>Accesses</h2>
                <Table aria-label={"accesses"}>
                    <TableHeader>
                        <TableColumn key={"frontendToken"}>Frontend Token</TableColumn>
                        <TableColumn key={"token"}>Token</TableColumn>
                        <TableColumn key={"bindAddress"}>Bind Address</TableColumn>
                        <TableColumn key={"responseHeaders"}>Response Headers</TableColumn>
                    </TableHeader>
                    <TableBody>
                        { status.accesses ? status.accesses.map((r) =>
                                <TableRow>
                                    <TableCell>{r.frontendToken}</TableCell>
                                    <TableCell>{r.token}</TableCell>
                                    <TableCell>{r.bindAddress}</TableCell>
                                    <TableCell>{r.responseHeaders}</TableCell>
                                </TableRow>
                            ) : null
                        }
                    </TableBody>
                </Table>
            </div>
            <div>
                <h2>Shares</h2>
                <Table aria-label="shares">
                    <TableHeader>
                        <TableColumn key="token">Token</TableColumn>
                        <TableColumn key={"reserved"}>Reserved</TableColumn>
                        <TableColumn key="shareMode">Share Mode</TableColumn>
                        <TableColumn key="backendMode">Backend Mode</TableColumn>
                        <TableColumn key="backendEndpoint">Target</TableColumn>
                        <TableColumn key="closed">Closed</TableColumn>
                        <TableColumn key="frontendEndpoints">Frontend Endpoints</TableColumn>
                    </TableHeader>
                    <TableBody>
                        { status.shares ? status.shares.map((r) =>
                                <TableRow>
                                    <TableCell>{r.token}</TableCell>
                                    <TableCell>{''+r.reserved}</TableCell>
                                    <TableCell>{r.shareMode}</TableCell>
                                    <TableCell>{r.backendMode}</TableCell>
                                    <TableCell>{r.backendEndpoint}</TableCell>
                                    <TableCell>{''+r.closed}</TableCell>
                                    <TableCell>{r.shareMode === 'public' ? r.frontendEndpoints : 'N/A'}</TableCell>
                                </TableRow>
                            ) : null
                        }
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
