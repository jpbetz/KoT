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
		api.getModules(this.onLoaded.bind(this))
	}

	onLoaded(data) {
		this.setState({modules: data});
	}

	render() {
		const { classes } = this.props;
		let pressureMarks = [{value: 1, label: "1 bar"}, {value: 10, label: "10 bar"}, {value: 15, label: "15 bar"}];
		let pumpMarks = [{value: 0, label: "0"}, {value: 5, label: "5"}];

		if (!this.state.modules) {
			return (
					<div>Loading</div>
			);
		}

		return (
				<div>
					<Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
						Auqanetes Deep Sea Research Station
					</Typography>
					<Container maxWidth="md" component="main">
						<Grid container alignItems="flex-end" spacing={5}>
							{this.state.modules.map(module => (
									<Grid item key={module.id} xs={12} sm={6} md={4}>
										<Card>
											<CardHeader
													title={titleCase(module.id + " Module")}
													titleTypographyProps={{align: 'center'}}
													className={classes.cardHeader}
											/>
											<CardContent>
												<div className={classes.cardContent}>

													<Typography component="h2" variant="h3" color="textPrimary">
														<Value fractionDigits={2} path={module.id + "." + module.pressureSensor.id + ".pressure"}/>
													</Typography>
													<Typography variant="h6" color="textSecondary">
														bar
													</Typography>
												</div>
												<ul>
													<Typography component="li" variant="subtitle1" align="center" key={"pump"}>
														Pumps Running: <Value path={module.id + "." + module.pump.id + ".activeCount"}/>
													</Typography>

													<Typography component="li" variant="subtitle1" align="center" key={"alarm"}>
														Alarm: <Light path={module.id + "." + module.waterAlarm.id + ".alarm"}/>
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
															Pressure Override
															<Slider min={1} max={15} marks={pressureMarks}
																			path={module.id + "." + module.pressureSensor.id + ".pressure"}/>
														</p>
														<p>
															Pump Override
															<Slider min={0} max={5} marks={pumpMarks} input path={module.id + "." + module.pump.id + ".activeCount"}/>
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
