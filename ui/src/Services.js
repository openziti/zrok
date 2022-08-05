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
		},
		{
			name: 'Active',
			selector: row => row.active,
			sortable: true
		}
	]

	const conditionalRowStyles = [
		{
			when: row => !row.active,
			style: {
				display: 'none'
			}
		}
	]

	return (
		<div className={"nested-services"}>
			{ props.services && props.services.length > 0 && (
				<DataTable
					columns={columns}
					data={props.services}
					defaultSortFieldId={1}
					conditionalRowStyles={conditionalRowStyles}
				/>
			)}
		</div>
	)
}

export default Services;