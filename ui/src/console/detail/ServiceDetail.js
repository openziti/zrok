import * as metadata from "../../api/metadata";
import {Sparklines, SparklinesLine, SparklinesSpots} from "react-sparklines";
import {useEffect, useState} from "react";

const ServiceDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadata.getServiceDetail(props.selection.id)
            .then(resp => {
               setDetail(resp.data);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getServiceDetail(props.selection.id)
                .then(resp => {
                    setDetail(resp.data);
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.selection]);

    if(detail) {
        return (
            <div>
                <h2>Service: {detail.token}</h2>
                <div className={"zrok-big-sparkline"}>
                    <Sparklines data={detail.metrics} limit={60} height={20}>
                        <SparklinesLine color={"#3b2693"} />
                        <SparklinesSpots />
                    </Sparklines>
                </div>
            </div>
        );
    }
    return <></>;
}

export default ServiceDetail;