import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import buildOverview from "./model/overview.js";
import NavBar from "./NavBar.jsx";
import {Box, Button, MenuItem, Modal, TextField} from "@mui/material";
import {useFormik} from "formik";

const AgentUi = () => {
    const [version, setVersion] = useState("");
    const [overview, setOverview] = useState(new Map());
    const [newShare, setNewShare] = useState(false);
    const [newAccess, setNewAccess] = useState(false);

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));
    const openNewShare = () => {
        setNewShare(true);
    }
    const closeNewShare = () => {
        setNewShare(false);
    }
    const openNewAccess = () => {
        setNewAccess(true);
    }
    const closeNewAccess = () => {
        setNewAccess(false);
    }
    const modalStyle = {
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
                    setOverview(buildOverview(data));
                }
            });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    });

    const releaseShare = (opts) => {
        api.agentReleaseShare(opts, (err, data) => {
            console.log(data);
        });
    }

    const releaseAccess = (opts) => {
        api.agentReleaseAccess(opts, (err, data) => {
            console.log(data);
        });
    }

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

    const shareHandler = (values) => {
        switch(values.shareMode) {
            case "public":
                api.agentSharePublic({
                    target: values.target,
                    backendMode: values.backendMode,
                }, (err, data) => {
                    closeNewShare();
                });
                break;

            case "private":
                api.agentSharePrivate({
                    target: values.target,
                    backendMode: values.backendMode,
                }, (err, data) => {
                    closeNewShare();
                });
                break;
        }
    }

    const accessHandler = (values) => {
        api.agentAccessPrivate({
            token: values.token,
            bindAddress: values.bindAddress,
        }, (err, data) => {
            closeNewAccess();
        });
    }

    const newShareForm = useFormik({
        initialValues: {
            shareMode: "public",
            backendMode: "proxy",
            target: "",
        },
        onSubmit: shareHandler,
    });

    const newAccessForm = useFormik({
        initialValues: {
            token: "",
            bindAddress: "",
        },
        onSubmit: accessHandler,
    })

    return (
        <>
            <NavBar version={version} shareClick={openNewShare} accessClick={openNewAccess} />
            <RouterProvider router={router} />

            <Modal
                open={newShare}
                onClose={closeNewShare}
            >
                <Box sx={{ ...modalStyle }}>
                    <h2>Share...</h2>
                    <form onSubmit={newShareForm.handleSubmit}>
                        <TextField
                            fullWidth
                            select
                            id="shareMode"
                            name="shareMode"
                            label="Share Mode"
                            value={newShareForm.values.shareMode}
                            onChange={newShareForm.handleChange}
                            onBlur={newShareForm.handleBlur}
                            sx={{ mt: 2 }}
                        >
                            <MenuItem value="public">public</MenuItem>
                            <MenuItem value="private">private</MenuItem>
                        </TextField>
                        <TextField
                            fullWidth select
                            id="backendMode"
                            name="backendMode"
                            label="Backend Mode"
                            value={newShareForm.values.backendMode}
                            onChange={newShareForm.handleChange}
                            onBlur={newShareForm.handleBlur}
                            sx={{ mt: 2 }}
                        >
                            <MenuItem value="proxy">proxy</MenuItem>
                            <MenuItem value="web">web</MenuItem>
                        </TextField>
                        <TextField
                            fullWidth
                            id="target"
                            name="target"
                            label="Target"
                            value={newShareForm.values.target}
                            onChange={newShareForm.handleChange}
                            onBlur={newShareForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                        <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Create Share</Button>
                    </form>
                </Box>
            </Modal>
            <Modal
                open={newAccess}
                onClose={closeNewAccess}
            >
                <Box sx={{...modalStyle}}>
                    <h2>Access...</h2>
                    <form onSubmit={newAccessForm.handleSubmit}>
                        <TextField
                            fullWidth
                            id="token"
                            name="token"
                            label="Share Token"
                            value={newAccessForm.values.token}
                            onChange={newAccessForm.handleChange}
                            onBlur={newAccessForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                        <TextField
                            fullWidth
                            id="bindAddress"
                            name="bindAddress"
                            label="Bind Address"
                            value={newAccessForm.values.bindAddress}
                            onChange={newAccessForm.handleChange}
                            onBlur={newAccessForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                        <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Create Access</Button>
                    </form>
                </Box>
            </Modal>
        </>
);
}

export default AgentUi;