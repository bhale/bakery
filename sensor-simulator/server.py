from flask import Flask
from random import randint, choice
import json
from time import time
import copy

app = Flask(__name__)

numSensors = 1000

# this is going to generate a list of fake sensors, runs at startup
def createSensors():
    sensors = [] 
    for x in range(0, numSensors):

	sensorType = choice(['Temperature', 'Wattage'])

	if sensorType == "Temperature":	
		unit = "Degrees F"

	if sensorType == "Wattage":
		unit = "Watts"

        sensors.append({'id': x, 'name': sensorType, 'unit': unit})

    return sensors

# generate random values appropriate to the sensor type, runs at collection
def getSensorValues(sensors):
    sensorValues = copy.copy(sensors)
    for sensor in sensorValues:
        if sensor['name'] == "Temperature":
            sensor['value'] = randint(40,80)
        if sensor['name'] == "Wattage":
            sensor['value'] = randint(70,500)
    return sensorValues

sensors = createSensors()

@app.route("/")
def index():
    return json.dumps(getSensorValues(sensors))

@app.route("/sensors")
def getSensors():
    return json.dumps(getSensorValues(sensors))

# current value for a single sensor by id
@app.route('/sensors/<sensorId>')
def getSensor(sensorId):
    sensorId = int(sensorId)
    thisSensor = []
    thisSensor.append(sensors[sensorId]) 
    values = getSensorValues(thisSensor) 
    return json.dumps(values)

if __name__ == "__main__":
    app.debug = False
    app.run()
