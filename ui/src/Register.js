import { useParams } from 'react-router-dom';

const Register = () => {
    const { token } = useParams();

    return (
        <div className={"zrok"}>
            <div className={"container"}>
                <div className={"header"}>
                    <img alt={"ziggy goes to space"} src="/ziggy.svg" width={"65px"} />
                    <p className={"header-title"}>zrok</p>
                </div>
                <div className={"main"}>
                    <h1>Register a new zrok account!</h1>
                </div>
            </div>
        </div>
    )
}

export default Register;