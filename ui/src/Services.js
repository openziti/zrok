import DataTable from 'react-data-table-component';
import {useEffect} from "react";

const Services = (props) => {
	useEffect((props) => {
		console.log(props)
	}, [])

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