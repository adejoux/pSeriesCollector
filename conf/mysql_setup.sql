create database pseriescollector;
GRANT USAGE ON `pseriescollector`.* to 'pseriescollectoruser'@'localhost' identified by 'pseriescollectorpass';
GRANT ALL PRIVILEGES ON `pseriescollector`.* to 'pseriescollectoruser'@'localhost' with grant option;
flush privileges;
