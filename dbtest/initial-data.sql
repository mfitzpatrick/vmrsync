-- SQL file for configuring some testing initial data with the
-- VMRMEMBERS Firebird database

CONNECT 'VMRMEMBERS.FDB';

INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('GREEN', 'Y');
INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('WHITE', 'Y');
INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('BLACK', 'Y');
INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('BLUE', 'Y');
INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('RED', 'Y');
INSERT INTO CREWS (CREWNAME,BOATCREW) VALUES ('YELLOW', 'Y');

INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Recruit',1,'R');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Crew',2,'CR');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Senior Crew',3,'SC');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Inshore Skipper',10,'IS');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Endorsed Inshore Skipper',11,'EIS');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Coxswain',12,'C');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Offshore Skipper',13,'OS');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Duty Skipper',20,'DS');
INSERT INTO RANKS (RANKNAME,RANKING,RANKNAMESHORT) VALUES ('Senior Skipper',21,'SS');

INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (1,'Fudd','Elmer','elmer.fudd@mrq.org.au','WHITE');
INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (2,'Martian','Marvin','marvin.the.martian@mrq.org.au','WHITE');
INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (3,'Bunny','Bugs','bugs.bunny@mrq.org.au','WHITE');
INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (4,'Bird','Tweety','tweety.bird@mrq.org.au','WHITE');
INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (5,'Devil','Tasmanian','tasmanian.devil@mrq.org.au','WHITE');
INSERT INTO MEMBERS (MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,CURRENTCREW) VALUES
    (6,'Pig','Porky','porky.pig@mrq.org.au','GREEN');

INSERT INTO RANKHISTORY (MEMBERNO,RANKDATE,RANKNAME) VALUES
    (1,'2019-06-28','OS');
INSERT INTO RANKHISTORY (MEMBERNO,RANKDATE,RANKNAME) VALUES
    (2,'2020-03-07','OS');
INSERT INTO RANKHISTORY (MEMBERNO,RANKDATE,RANKNAME) VALUES
    (3,'2018-01-07','SC');
INSERT INTO RANKHISTORY (MEMBERNO,RANKDATE,RANKNAME) VALUES
    (4,'2020-03-07','SC');
INSERT INTO RANKHISTORY (MEMBERNO,RANKDATE,RANKNAME) VALUES
    (5,'2022-01-25','CR');

INSERT INTO VESSELS (VESSELNO,VESSELNAMESHORT) VALUES (1,'MR1');
INSERT INTO VESSELS (VESSELNO,VESSELNAMESHORT) VALUES (2,'MR2');
INSERT INTO VESSELS (VESSELNO,VESSELNAMESHORT) VALUES (3,'MR4');
INSERT INTO VESSELS (VESSELNO,VESSELNAMESHORT) VALUES (4,'MR5');

-- Duty log entry for a given crew shift
INSERT INTO DUTYLOG (DUTYSEQUENCE,DUTYDATE,CREW,SKIPPER) VALUES (1,'2022-01-02','GREEN',6);
INSERT INTO DUTYLOG (DUTYSEQUENCE,DUTYDATE,CREW,SKIPPER) VALUES (2,'2022-01-03','WHITE',1);

-- Voyage information
INSERT INTO DUTYJOBS (JOBDUTYSEQUENCE,JOBJOBSEQUENCE,JOBTIMEOUT,JOBTIMEIN,JOBDUTYVESSELNAME,JOBDUTYVESSELNO,
        JOBTYPE,JOBACTIONTAKEN) VALUES
    (1,1,'2022-01-01 06:00:35','2020-01-01 08:25:42','MR2',2,'TRAINING','TRAINING');
INSERT INTO DUTYJOBS (JOBDUTYSEQUENCE,JOBJOBSEQUENCE,JOBTIMEOUT,JOBTIMEIN,JOBDUTYVESSELNAME,JOBDUTYVESSELNO,
        JOBTYPE,JOBACTIONTAKEN) VALUES
    (2,2,'2022-01-01 09:10:00','2020-01-01 10:00:22','MR5',4,'TRAINING','TRAINING');

-- VMR crew onboard vessel for a given DUTYJOBS entry
INSERT INTO DUTYJOBSCREW (CREWDUTYSEQUENCE,CREWJOBSEQUENCE,CREWMEMBER,CREWRANKING,SKIPPER,CREWONJOB) VALUES
    (2,1,1,13,'Y','Y');
INSERT INTO DUTYJOBSCREW (CREWDUTYSEQUENCE,CREWJOBSEQUENCE,CREWMEMBER,CREWRANKING,SKIPPER,CREWONJOB) VALUES
    (2,1,3,3,'N','Y');
INSERT INTO DUTYJOBSCREW (CREWDUTYSEQUENCE,CREWJOBSEQUENCE,CREWMEMBER,CREWRANKING,SKIPPER,CREWONJOB) VALUES
    (2,2,2,13,'Y','Y');
INSERT INTO DUTYJOBSCREW (CREWDUTYSEQUENCE,CREWJOBSEQUENCE,CREWMEMBER,CREWRANKING,SKIPPER,CREWONJOB) VALUES
    (2,2,5,2,'N','Y');

