import {useEffect, useState} from "react";
import {GetAgentApi} from "./model/api.ts";
import NavBar from "./NavBar.tsx";
import {AgentObject, buildOverview} from "./model/overview.ts";
import Overview from "./Overview.tsx";
import NewShareModal from "./NewShareModal.tsx";
import NewAccessModal from "./NewAccessModal.tsx";

const AgentUi = () => {
    const [version, setVersion] = useState("unset");
    const [overview, setOverview] = useState(new Array<AgentObject>());
    const [newShareOpen, setNewShareOpen] = useState(false);
    const [newAccessOpen, setNewAccessOpen] = useState(false);

    const openNewShare = () => {
        setNewShareOpen(true);
    }
    const closeNewShare = () => {
        setNewShareOpen(false);
    }

    const openNewAccess = () => {
        setNewAccessOpen(true);
    }
    const closeNewAccess = () => {
        setNewAccessOpen(false);
    }

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
            <NavBar shareClick={openNewShare} accessClick={openNewAccess} />
            <Overview overview={overview} shareClick={openNewShare} accessClick={openNewAccess} />
            <NewShareModal isOpen={newShareOpen} close={closeNewShare} />
            <NewAccessModal isOpen={newAccessOpen} close={closeNewAccess} />
        </>
    );
}

export default AgentUi;