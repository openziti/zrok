import DataTable from 'react-data-table-component';
import {Sparklines, SparklinesLine, SparklinesSpots} from 'react-sparklines';
import {mdiCloseOutline} from "@mdi/js";
import Icon from "@mdi/react";

const Services = (props) => {
	const humanizeDuration = require("humanize-duration")
	const untunnelService = (svcName) => {
		if(window.confirm('really disable service "' + svcName +'"?')) {
			console.log("will disable serivce: " + svcName)
		}
	}

	const columns = [
		{
			name: 'Frontend',
			selector: row => row.frontend,
			sortable: true,
		},
		{
			name: 'Backend',
			selector: row => row.backend,
			sortable: true,
		},
		{
			name: 'Actions',
			selector: row => <>
				<button data-value={row.name} onClick={e => untunnelService(row.name)} title={"Disable Service '"+row.name+"'"}>
					<Icon path={mdiCloseOutline} size={0.7}/>
				</button>
			</>
		},
		{
			name: 'Uptime',
			selector: row => humanizeDuration(Date.now() - row.updatedAt),
		},
		{
			name: 'Activity',
			cell: row => <Sparklines data={row.metrics} height={20} limit={60}><SparklinesLine color={"#3b2693"}/><SparklinesSpots/></Sparklines>
		}
	]

	return (
		<div className={"nested-services"}>
			{ props.services && props.services.length > 0 && (
				<DataTable
					columns={columns}
					data={props.services}
					defaultSortFieldId={1}
				/>
			)}
		</div>
	)
}

export default Services;