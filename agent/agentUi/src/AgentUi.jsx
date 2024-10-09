import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import buildOverview from "./model/overview.js";
import NavBar from "./NavBar.jsx";
import {Box, Modal} from "@mui/material";

const AgentUi = () => {
    const [version, setVersion] = useState("");
    const [shares, setShares] = useState([]);
    const [accesses, setAccesses] = useState([]);
    const [overview, setOverview] = useState(new Map());
    const [newShare, setNewShare] = useState(false);

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));
    const openNewShare = () => {
        setNewShare(true);
    }
    const closeNewShare = () => {
        setNewShare(false);
    }
    const shareStyle = {
        position: 'absolute',
        top: '25%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        width: 600,
        bgcolor: 'background.paper',
        boxShadow: 24,
        p: 4,
    };

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
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                if(mounted) {
                    setShares(data.shares);
                    setAccesses(data.accesses);
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
            element: <Overview version={version} overview={overview} />
        },
        {
            path: "/share/:token",
            element: <ShareDetail version={version} shares={shares} />
        }
    ]);

    return (
        <>
            <NavBar version={version} shareClick={openNewShare} />
            <RouterProvider router={router} />
            <Modal
                open={newShare}
                onClose={closeNewShare}
            >
                <Box sx={{ ...shareStyle }}>
                    <h2>New Share</h2>
                </Box>
            </Modal>
        </>
    );
}

export default AgentUi;