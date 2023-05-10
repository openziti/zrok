import {mdiAccountBox} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Col, Container, Row, Tab, Tabs, Tooltip} from "react-bootstrap";
import SecretToggle from "../../SecretToggle";
import React, {useEffect, useState} from "react";
import * as metadata from "../../../api/metadata";
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, XAxis, YAxis } from "recharts";
import moment from "moment";

const AccountDetail = (props) => {
    const customProperties = {
        token: row => <SecretToggle secret={row.value} />
    }

    return (
        <div>
            <h2><Icon path={mdiAccountBox} size={2} />{" "}{props.user.email}</h2>
            <Tabs defaultActiveKey={"detail"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <PropertyTable object={props.user} custom={customProperties}/>
                </Tab>
                <Tab eventKey={"metrics"} title={"Metrics"}>
                    <MetricsTab />
                </Tab>
            </Tabs>
        </div>
    );
}

const MetricsTab = (props) => {
    const [metrics30, setMetrics30] = useState(buildMetrics([]));
    const [metrics7, setMetrics7] = useState(buildMetrics([]));
    const [metrics1, setMetrics1] = useState(buildMetrics([]));

    useEffect(() => {
        metadata.getAccountMetrics()
            .then(resp => {
                setMetrics30(buildMetrics(resp.data));
            });
        metadata.getAccountMetrics({duration: "168h"})
            .then(resp => {
                setMetrics7(buildMetrics(resp.data));
            });
        metadata.getAccountMetrics({duration: "24h"})
            .then(resp => {
                setMetrics1(buildMetrics(resp.data));
            });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getAccountMetrics()
                .then(resp => {
                    if(mounted) {
                        setMetrics30(buildMetrics(resp.data));
                    }
                });
            metadata.getAccountMetrics({duration: "168h"})
                .then(resp => {
                    setMetrics7(buildMetrics(resp.data));
                });
            metadata.getAccountMetrics({duration: "24h"})
                .then(resp => {
                    setMetrics1(buildMetrics(resp.data));
                });
        }, 5000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

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

const buildMetrics = (m) => {
    let metrics = {
        data: m.samples,
        rx: 0,
        tx: 0
    }
    if(m.samples) {
        m.samples.forEach(s => {
            metrics.rx += s.rx;
            metrics.tx += s.tx;
        });
    }
    return metrics;
}

const bytesToSize = (sz) => {
    let absSz = sz;
    if(absSz < 0) {
        absSz *= -1;
    }
    const unit = 1000
    if(absSz < unit) {
        return '' + absSz + ' B';
    }
    let div = unit
    let exp = 0
    for(let n = absSz / unit; n >= unit; n /= unit) {
        div *= unit;
        exp++;
    }

    return '' + (sz / div).toFixed(1) + ' ' + "kMGTPE"[exp] + 'B';
}

export default AccountDetail;