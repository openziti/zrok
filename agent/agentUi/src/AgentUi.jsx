import {useEffect, useState} from "react";
import NavBar from "./NavBar.jsx";
import {getAgentApi} from "./model/handler.js";
import {buildOverview} from "./model/overview.js";
import Overview from "./Overview.jsx";
import NewAccessModal from "./NewAccessModal.jsx";
import NewShareModal from "./NewShareModal.jsx";

const AgentUi = () => {
    const [version, setVersion] = useState("");
    const [overview, setOverview] = useState([]);
    const [newAccessOpen, setNewAccessOpen] = useState(false);
    const [newShareOpen, setNewShareOpen] = useState(false);

    const openNewAccess = () => {
        setNewAccessOpen(true);
    }
    const closeNewAccess = () => {
        setNewAccessOpen(false);
    }

    const openNewShare = () => {
        setNewShareOpen(true);
    }
    const closeNewShare = () => {
        setNewShareOpen(false);
    }

    useEffect(() => {
        getAgentApi().agentVersion((e, d) => {
            setVersion(d.v);
        });
        return () => {
            setVersion("");
        }
    }, []);

    useEffect(() => {
        let interval = setInterval(() => {
            getAgentApi().agentStatus((e, d) => {
                if(e) {
                    setOverview([]);
                    console.log("agentStatus", e);
                } else {
                    setOverview(buildOverview(d));
                }
            });
        }, 1000);
        return () => {
            clearInterval(interval);
            setOverview([]);
        }
    }, []);


    return (
        <>
            <NavBar version={version} shareClick={openNewShare} accessClick={openNewAccess} />
            <Overview overview={overview} />
            <NewAccessModal isOpen={newAccessOpen} close={closeNewAccess} />
            <NewShareModal isOpen={newShareOpen} close={closeNewShare} />
        </>
    );
}

export default AgentUi;