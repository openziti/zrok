import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import buildOverview from "./model/overview.js";
import NavBar from "./NavBar.jsx";
import NewShareModal from "./NewShareModal.jsx";
import NewAccessModal from "./NewAccessModal.jsx";
import {accessHandler, getAgentApi, releaseAccess, releaseShare, shareHandler} from "./model/handler.js";

const AgentUi = () => {
    const [version, setVersion] = useState("");
    const [overview, setOverview] = useState([]);

    const [newShare, setNewShare] = useState(false);
    const openNewShare = () => {
        setNewShare(true);
    }
    const closeNewShare = () => {
        setNewShare(false);
    }

    const [newAccess, setNewAccess] = useState(false);
    const openNewAccess = () => {
        setNewAccess(true);
    }
    const closeNewAccess = () => {
        setNewAccess(false);
    }

    useEffect(() => {
        getAgentApi().agentVersion((err, data) => {
            setVersion(data.v);
        });
        return () => {
            setVersion("");
        }
    }, []);

    const newStatus = (err, data) => {
        if(err) {
            console.log("newState", err);
            setOverview([]);
        } else {
            setOverview(buildOverview(data));
        }
    }

    useEffect(() => {
        let interval = setInterval(() => {
            getAgentApi().agentStatus(newStatus);
        }, 1000);
        return () => {
            clearInterval(interval);
            setOverview([]);
        }
    }, []);

    const router = createBrowserRouter([
        {
            path: "/",
            element: <Overview
                releaseShare={releaseShare}
                releaseAccess={releaseAccess}
                version={version}
                overview={overview}
                shareClick={openNewShare}
                accessClick={openNewAccess}
            />
        },
        {
            path: "/share/:token",
            element: <ShareDetail version={version} />
        }
    ]);

    return (
        <>
            <NavBar version={version} shareClick={openNewShare} accessClick={openNewAccess} />
            <RouterProvider router={router} />
            <NewShareModal show={newShare} close={closeNewShare} handler={shareHandler} />
            <NewAccessModal show={newAccess} close={closeNewAccess} handler={accessHandler} />
        </>
    );
}

export default AgentUi;