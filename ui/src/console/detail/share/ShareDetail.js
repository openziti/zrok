import * as metadata from "../../../api/metadata";
import {Sparklines, SparklinesLine, SparklinesSpots} from "react-sparklines";
import {useEffect, useState} from "react";
import {mdiEyeOffOutline, mdiEyeOutline, mdiShareVariant} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import {secretString} from "../../Utils";
import ActionsTab from "./ActionsTab";

const ShareDetail = (props) => {
    const [detail, setDetail] = useState({});
    const [showZId, setShowZId] = useState(false);

    useEffect(() => {
        metadata.getShareDetail(props.selection.id)
            .then(resp => {
                let detail = resp.data;
                detail.envZId = props.selection.envZId;
                setDetail(detail);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getShareDetail(props.selection.id)
                .then(resp => {
                    let detail = resp.data;
                    detail.envZId = props.selection.envZId;
                    setDetail(detail);
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.selection]);

    const customProperties = {
        metrics: row => (
            <Sparklines data={row.value} limit={60} height={10}>
                <SparklinesLine color={"#3b2693"}/>
                <SparklinesSpots/>
            </Sparklines>
        ),
        frontendEndpoint: row => (
            <a href={row.value} target="_">{row.value}</a>
        ),
        backendProxyEndpoint: row => {
            if(row.value.startsWith("http://") || row.value.startsWith("https://")) {
                return <a href={row.value} target="_">{row.value}</a>;
            }
            return row.value;
        },
        zId: row => {
            if(showZId) {
                return <span>{row.value} <Icon path={mdiEyeOffOutline} size={0.7} onClick={() => setShowZId(false)} /></span>
            } else {
                return <span>{secretString(row.value)} <Icon path={mdiEyeOutline} size={0.7} onClick={() => setShowZId(true)} /></span>
            }
        }
    }

    if(detail) {
        return (
            <div>
                <h2><Icon path={mdiShareVariant} size={2} />{" "}{detail.backendProxyEndpoint}</h2>
                <Tabs defaultActiveKey={"detail"}>
                    <Tab eventKey={"detail"} title={"Detail"}>
                        <PropertyTable object={detail} custom={customProperties} />
                    </Tab>
                    <Tab eventKey={"actions"} title={"Actions"}>
                        <ActionsTab share={detail} />
                    </Tab>
                </Tabs>
            </div>
        );
    }
    return <></>;
}

export default ShareDetail;