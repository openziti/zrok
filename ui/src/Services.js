import DataTable from 'react-data-table-component';
import {useEffect} from "react";

const Services = (props) => {
	useEffect((props) => {
		console.log(props)
	}, [])

	const columns = [
		{
			name: 'Endpoint',
			selector: row => row.endpoint,
			sortable: true,
		},
		{
			name: 'Service Id',
			selector: row => row.zitiServiceId,
			sortable: true,
		}
	]

	return (
		<div className={"nested-services"}>
			{ props.services && props.services.length > 0 && (
				<DataTable columns={columns} data={props.services} defaultSortFieldId={1}/>
			)}
		</div>
	)
}

export default Services;