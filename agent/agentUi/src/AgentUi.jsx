import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import buildOverview from "./model/overview.js";
import NavBar from "./NavBar.jsx";
import NewShareModal from "./NewShareModal.jsx";
import NewAccessModal from "./NewAccessModal.jsx";
import {accessHandler, releaseAccess, releaseShare, shareHandler} from "./model/handler.js";

const AgentUi = () => {
    const [version, setVersion] = useState("");
    const [overview, setOverview] = useState(new Map());

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

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
            if(mounted) {
                setVersion(data.v);
            }
        });
    }, []);

    useEffect(() => {
        let mounted = true;
        api.agentStatus((err, data) => {
            if(mounted) {
                setOverview(buildOverview(data));
            }
        });
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                if(mounted) {
                    setOverview(buildOverview(data));
                }
            });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    });

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