import React from 'react';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import Container from '@material-ui/core/Container';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';
import { withStyles } from '@material-ui/styles';
import Value from "./Value.js";
import Slider from "./Slider.js";
import Switch from "./Switch.js";
import Light from "./Light.js";
import './App.css';
import * as api from "./api";

const styles = theme => ({
	'@global': {
		body: {
			backgroundColor: theme.palette.common.white,
		},
		ul: {
			margin: 0,
			padding: 0,
		},
		li: {
			listStyle: 'none',
		},
	},
	cardHeader: {
		backgroundColor: theme.palette.grey[300],
	},
	cardSplit: {
		backgroundColor: theme.palette.grey[100],
	},
	cardContent: {
		display: 'flex',
		justifyContent: 'center',
		alignItems: 'baseline',
		marginBottom: theme.spacing(2),
	},
});

function titleCase(word) {
	return word.charAt(0).toUpperCase()+ word.slice(1);
}

class App extends React.Component {
	constructor(props) {
		super(props);
		this.state = {sensor: {value: 0}};
		this.path = this.props.path
	}

	componentDidMount() {
		api.getDataset(this.onLoaded.bind(this));
		api.addOnModuleChangedListener(this.onModulesChanged.bind(this));
	}

	componentWillUnmount() {
		api.removeOnModuleChangedListener(this.onModulesChanged.bind(this));
	}

	onModulesChanged() {
		api.getDataset(this.onLoaded.bind(this));
	}

	onLoaded(data) {
		this.setState({modules: data.modules, devices: data.devices});
	}

	render() {
		const { classes } = this.props;
		let pressureMarks = [{value: 1, label: "1"}, {value: 10, label: "10"}, {value: 15, label: "15"}];
		let pumpMarks = [{value: 0, label: "0"}, {value: 5, label: "5"}];

		if (!this.state.modules) {
			return (
					<div>Loading</div>
			);
		}

		return (
				<div>
					<Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
						Deep Sea Research Station
					</Typography>
					<Container maxWidth="md" component="main">
						<Grid container alignItems="flex-end" spacing={5}>
							{Object.values(this.state.modules).map(module => (
									<Grid item key={module.metadata.name} xs={12} sm={6} md={4}>
										<Card>
											<CardHeader
													title={titleCase(module.metadata.name + " Module")}
													titleTypographyProps={{align: 'center'}}
													className={classes.cardHeader}
											/>
											<CardContent>
												<div className={classes.cardContent}>

													<Typography component="h2" variant="h3" color="textPrimary">
														<Value fractionDigits={2} path={module.metadata.name + "." + this.state.devices[module.spec.devices.pressureSensor].metadata.name + ".pressure"}/>
													</Typography>
													<Typography variant="h6" color="textSecondary">
														bar
													</Typography>
												</div>
												<ul>
													<Typography component="li" variant="subtitle1" align="center" key={"pump"}>
														<Value path={module.metadata.name + "." + this.state.devices[module.spec.devices.pump].metadata.name + ".activeCount"}/> Pumps Running
													</Typography>

													<Typography component="li" variant="subtitle1" align="center" key={"alarm"}>
														<Light path={module.metadata.name + "." + this.state.devices[module.spec.devices.waterAlarm].metadata.name + ".alarm"}/>
													</Typography>
												</ul>
											</CardContent>
											<CardHeader
													title={"Simulation Controls"}
													titleTypographyProps={{align: 'center'}}
													className={classes.cardSplit}
											/>
											<CardContent>
												<div className={classes.cardContent}>
													<ul>
														<p>
															Pressure
															<Slider min={1} max={15} marks={pressureMarks}
																			path={module.metadata.name + "." + this.state.devices[module.spec.devices.pressureSensor].metadata.name + ".pressure"}/>
														</p>
														<p>
															Pumps
															<Slider min={0} max={5} marks={pumpMarks} input path={module.metadata.name + "." + this.state.devices[module.spec.devices.pump].metadata.name + ".activeCount"}/>
														</p>
														<p>
															Alarm
															<Switch path={module.metadata.name + "." + this.state.devices[module.spec.devices.waterAlarm].metadata.name + ".alarm"}/>
														</p>
													</ul>
												</div>
											</CardContent>
										</Card>
									</Grid>
							))}
						</Grid>
					</Container>
				</div>
		);
	}
}

export default withStyles(styles)(App);
