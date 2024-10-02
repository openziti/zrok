"use client";

import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "@/api/src";

export default function Home() {
    const [version, setVersion] = useState("");
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

    return (
        <div className="grid grid-rows-[20px_1fr_20px] min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
            <p>Agent: {version}</p>
        </div>
    );
}
