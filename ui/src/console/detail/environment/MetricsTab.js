import React, {useEffect, useState} from "react";
import {buildMetrics} from "../../metrics/util";
import MetricsView from "../../metrics/MetricsView";
import { metadataApi } from "../../..";

const MetricsTab = (props) => {
	const [metrics30, setMetrics30] = useState(buildMetrics([]));
	const [metrics7, setMetrics7] = useState(buildMetrics([]));
	const [metrics1, setMetrics1] = useState(buildMetrics([]));

	useEffect(() => {
		metadataApi.getEnvironmentMetrics(props.selection.envZId)
			.then(resp => {
				setMetrics30(buildMetrics(resp));
			}).catch(err => {
				console.log(err)
			});
		metadataApi.getEnvironmentMetrics(props.selection.envZId, {duration: "168h"})
			.then(resp => {
				setMetrics7(buildMetrics(resp));
			}).catch(err => {
				console.log(err)
			});
		metadataApi.getEnvironmentMetrics(props.selection.envZId, {duration: "24h"})
			.then(resp => {
				setMetrics1(buildMetrics(resp));
			}).catch(err => {
				console.log(err)
			});
			// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [props.selection.id]);

	useEffect(() => {
		let mounted = true;
		let interval = setInterval(() => {
			metadataApi.getEnvironmentMetrics(props.selection.envZId)
				.then(resp => {
					if(mounted) {
						setMetrics30(buildMetrics(resp));
					}
				}).catch(err => {
					console.log(err)
				});
			metadataApi.getEnvironmentMetrics(props.selection.envZId, {duration: "168h"})
				.then(resp => {
					setMetrics7(buildMetrics(resp));
				}).catch(err => {
					console.log(err)
				});
			metadataApi.getEnvironmentMetrics(props.selection.envZId, {duration: "24h"})
				.then(resp => {
					setMetrics1(buildMetrics(resp));
				}).catch(err => {
					console.log(err)
				});
		}, 5000);
		return () => {
			mounted = false;
			clearInterval(interval);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [props.selection.id]);

	return (
		<MetricsView metrics30={metrics30} metrics7={metrics7} metrics1={metrics1} />
	);
};

export default MetricsTab;