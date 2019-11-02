import React from 'react';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import Container from '@material-ui/core/Container';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import Value from "./Value.js";
import Slider from "./Slider.js";
import Switch from "./Switch.js";
import Light from "./Light.js";
import './App.css';

const useStyles = makeStyles(theme => ({
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
}));

function App() {
	var pressureMarks = [{value:1, label:"1 bar"}, {value:10, label:"10 bar"}, {value:15, label:"15 bar"}];
	var pumpMarks = [{value:0, label:"0"}, {value:5, label:"5"}];
	var classes = useStyles();
	return (
		<div>
			<Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
	      Auqanetes Deep Sea Research Station
	    </Typography>
			<Container maxWidth="md" component="main">
	    <Grid container alignItems="flex-end" spacing={5}>
	    {["crew", "command", "research"].map(deviceID => (
				<Grid item key={deviceID} xs={12} sm={6} md={4}>
					<Card>
						<CardHeader
								title={deviceID}
								titleTypographyProps={{ align: 'center' }}
								className={classes.cardHeader}
						/>
						<CardContent>
							<div className={classes.cardContent}>

								<Typography component="h2" variant="h3" color="textPrimary">
									<Value deviceID={deviceID} ID="pressure"/>
								</Typography>
								<Typography variant="h6" color="textSecondary">
									bar
								</Typography>
							</div>
								<ul>
									<Typography component="li" variant="subtitle1" align="center" key={"pump"}>
										Pumps Active: <Value deviceID={deviceID} ID="pumpsActive" />
									</Typography>

									<Typography component="li" variant="subtitle1" align="center" key={"alarm"}>
										Alarm: <Light className={"alarm"} deviceID={deviceID} ID="alarm" />
									</Typography>
								</ul>
						</CardContent>
						<CardHeader
								title={"Simulation Controls"}
								titleTypographyProps={{ align: 'center' }}
								className={classes.cardSplit}
						/>
						<CardContent>
							<div className={classes.cardContent}>
								<ul>
								<p>
									Pressure
									<Slider min={1} max={15} marks={pressureMarks} deviceID={deviceID} ID="pressure" />
								</p>
								<p>
									Active Pumps
									<Slider min={0} max={5} marks={pumpMarks} input deviceID={deviceID} ID="pumpsActive" />
								</p>
								<p>
									Water Sensor
									<Switch deviceID={deviceID} ID="waterSensor" />
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

export default App;
