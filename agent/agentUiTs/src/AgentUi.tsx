import {useEffect, useState} from "react";
import {GetAgentApi} from "./model/api.ts";
import NavBar from "./NavBar.tsx";
import {AgentObject, buildOverview} from "./model/overview.ts";
import Overview from "./Overview.tsx";

const AgentUi = () => {
    const [version, setVersion] = useState("unset");
    const [overview, setOverview] = useState(new Array<AgentObject>());

    useEffect(() => {
		GetAgentApi().agentVersion()
            .then(r => {
                if(r.v) {
                    setVersion(r.v);
                } else {
                    console.log("unexpected", r);
                }
            })
            .catch(e => {
                console.log(e);
            });
    }, []);

    useEffect(() => {
        let interval = setInterval(() => {
            GetAgentApi().agentStatus()
                .then(r => {
                    setOverview(buildOverview(r));
                })
                .catch(e => {
                    console.log(e);
                })
        }, 1000);
        return () => {
            clearInterval(interval);
            setOverview(new Array<AgentObject>());
        }
    }, []);

    return (
        <>
            <NavBar version={version} />
            <Overview overview={overview} />
        </>
    );
}

export default AgentUi;