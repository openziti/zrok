import {Node} from "@xyflow/react";
import {Grid2, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import {Configuration, MetadataApi, Share} from "./api";
import {useEffect, useState} from "react";
import {User} from "./model/user.ts";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";

interface SharePanelProps {
    share: Node;
    user: User;
}

const SharePanel = ({ share, user }: SharePanelProps) => {
    const [detail, setDetail] = useState<Share>(null);

    const customProperties = {
        zId: row => <SecretToggle secret={row.value} />,
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString(),
        frontendEndpoint: row => <a href={row.value} target="_">{row.value}</a>
    }

    useEffect(() => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        console.log(share);
        let metadata = new MetadataApi(cfg);
        metadata.getShareDetail({ shrToken: share.data!.shrToken! as string })
            .then(d => {
                delete d.activity;
                delete d.limited;
                delete d.reserved;
                console.log("d", d);
                setDetail(d);
            })
            .catch(e => {
                console.log("SharePanel", e);
            })
    }, [share]);

    return (
        <Typography component="div">
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><ShareIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex" component="h3">{String(share.data.label)}</Grid2>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex">
                    <PropertyTable object={detail} custom={customProperties}/>
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default SharePanel;
