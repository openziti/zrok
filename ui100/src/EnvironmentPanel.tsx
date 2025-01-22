import {Node} from "@xyflow/react";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Tooltip, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import {useEffect, useRef, useState} from "react";
import {Configuration, Environment, EnvironmentApi, MetadataApi} from "./api";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import {modalStyle} from "./styling/theme.ts";

interface ReleaseEnvironmentProps {
    close: () => void;
    isOpen: boolean;
    detail: Environment;
    action: () => void;
}

const ReleaseEnvironmentModal = ({ close, isOpen, detail, action }: ReleaseEnvironmentProps) => {
    const [description, setDescription] = useState<String>("");
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>();
    checkedRef.current = checked;

    const toggleChecked = (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
    }, [isOpen]);

    useEffect(() => {
        if(detail && detail.description) {
            setDescription(detail.description);
        }
    }, [detail]);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Release Environment</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Would you like to release the environment <code>{description}</code> ?</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Releasing this environment will also release any shares and accesses that are associated with it.</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm the release of <code>{description}</code></p>} sx={{ mt: 2 }} />
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked} onClick={action}>Release</Button>
                </Grid2>
            </Box>
        </Modal>
    );
}

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({environment}: EnvironmentPanelProps) => {
    const user = useStore((state) => state.user);
    const [detail, setDetail] = useState<Environment>(null);
    const [releaseEnvironmentOpen, setReleaseEnvironmentOpen] = useState(false);

    const openReleaseEnvironment = () => {
        setReleaseEnvironmentOpen(true);
    }

    const closeReleaseEnvironment = () => {
        setReleaseEnvironmentOpen(false);
    }

    const customProperties = {
        zId: row => <SecretToggle secret={row.value} />,
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString()
    }

    const labels = {
        createdAt: "Created",
        updatedAt: "Updated"
    }

    useEffect(() => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let metadata = new MetadataApi(cfg);
        metadata.getEnvironmentDetail({envZId: environment.data!.envZId! as string})
            .then(d => {
                let env = d.environment!;
                delete env.activity;
                delete env.limited;
                delete env.zId;
                setDetail(env);
            })
            .catch(e => {
                console.log("EnvironmentPanel", e);
            })
    }, [environment]);

    const releaseEnvironment = () => {
        if(detail && detail.zId) {
            let cfg = new Configuration({
                headers: {
                    "X-TOKEN": user.token
                }
            });
            let environment = new EnvironmentApi(cfg);
            environment.disable({body: {identity: detail.zId}})
                .then(d => {
                    setReleaseEnvironmentOpen(false);
                })
                .catch(e => {
                    console.log("releaseEnvironment", e);
                });
        }
    }

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h3">{String(environment.data.label)}</Grid2>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Release Environment">
                        <Button variant="contained" color="error" onClick={openReleaseEnvironment}><DeleteIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                    </Grid2>
                </Grid2>
            </Typography>
            <ReleaseEnvironmentModal close={closeReleaseEnvironment} isOpen={releaseEnvironmentOpen} detail={detail} action={releaseEnvironment} />
        </>
    );
}

export default EnvironmentPanel;