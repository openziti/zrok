import React, {useEffect, useState} from "react";
import {buildMetrics, bytesToSize} from "../../metrics";
import * as metadata from "../../../api/metadata";
import {Col, Container, Row, Tooltip} from "react-bootstrap";
import {Bar, BarChart, CartesianGrid, ResponsiveContainer, XAxis, YAxis} from "recharts";
import moment from "moment";

const MetricsTab = (props) => {
    const [metrics30, setMetrics30] = useState(buildMetrics([]));
    const [metrics7, setMetrics7] = useState(buildMetrics([]));
    const [metrics1, setMetrics1] = useState(buildMetrics([]));

    useEffect(() => {
        console.log("token", props.share.token);
        metadata.getShareMetrics(props.share.token)
            .then(resp => {
                setMetrics30(buildMetrics(resp.data));
            });
        metadata.getShareMetrics(props.share.token, {duration: "168h"})
            .then(resp => {
                setMetrics7(buildMetrics(resp.data));
            });
        metadata.getShareMetrics(props.share.token, {duration: "24h"})
            .then(resp => {
                setMetrics1(buildMetrics(resp.data));
            });
    }, [props.share]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            console.log("token", props.share.token);
            metadata.getShareMetrics(props.share.token)
                .then(resp => {
                    if(mounted) {
                        setMetrics30(buildMetrics(resp.data));
                    }
                });
            metadata.getShareMetrics(props.share.token, {duration: "168h"})
                .then(resp => {
                    setMetrics7(buildMetrics(resp.data));
                });
            metadata.getShareMetrics(props.share.token, {duration: "24h"})
                .then(resp => {
                    setMetrics1(buildMetrics(resp.data));
                });
        }, 5000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.share]);

    return (
        <Container>
            <Row>
                <Col>
                    <h3>Last 30 Days:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(metrics30.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(metrics30.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={metrics30.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
            <Row>
                <Col>
                    <h3>Last 7 Days:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(metrics7.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(metrics7.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={metrics7.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
            <Row>
                <Col>
                    <h3>Last 24 Hours:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(metrics1.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(metrics1.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={metrics1.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
        </Container>
    );
}

export default MetricsTab;