-- MySQL dump 10.13  Distrib 5.6.25-73.1, for osx10.11 (x86_64)
--
-- Host: localhost    Database: httpDispacher
-- ------------------------------------------------------
-- Server version	5.6.25-73.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `DomainTable`
--

DROP TABLE IF EXISTS `DomainTable`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DomainTable` (
  `idDomainName` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `DomainName` char(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `NS` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`idDomainName`),
  UNIQUE KEY `DomainName_UNIQUE` (`DomainName`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `DomainTable`
--

LOCK TABLES `DomainTable` WRITE;
/*!40000 ALTER TABLE `DomainTable` DISABLE KEYS */;
INSERT INTO `DomainTable` VALUES (1,'api.weibo.cn.',NULL),(2,'weibo.cn.',NULL),(3,'sinaedge.com.',NULL);
/*!40000 ALTER TABLE `DomainTable` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RRTable`
--

DROP TABLE IF EXISTS `RRTable`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RRTable` (
  `idRRTable` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `RegionID` int(10) unsigned NOT NULL,
  `Rrtype` int(10) unsigned NOT NULL,
  `Class` int(10) unsigned NOT NULL,
  `Ttl` int(10) unsigned NOT NULL,
  `Target` varchar(45) NOT NULL,
  PRIMARY KEY (`idRRTable`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RRTable`
--

LOCK TABLES `RRTable` WRITE;
/*!40000 ALTER TABLE `RRTable` DISABLE KEYS */;
INSERT INTO `RRTable` VALUES (1,1,5,1,300,'weibo.cn'),(2,2,1,1,300,'180.149.153.216'),(3,2,1,1,300,'180.149.139.248');
/*!40000 ALTER TABLE `RRTable` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `RegionTable`
--

DROP TABLE IF EXISTS `RegionTable`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `RegionTable` (
  `idRegion` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `DomainNameID` int(10) unsigned NOT NULL,
  `StartIP` int(10) unsigned zerofill NOT NULL,
  `EndIP` int(10) unsigned zerofill NOT NULL,
  `NetAddr` int(10) unsigned NOT NULL,
  `NetMask` int(10) unsigned NOT NULL,
  PRIMARY KEY (`idRegion`),
  KEY `idx_union` (`StartIP`,`EndIP`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `RegionTable`
--

LOCK TABLES `RegionTable` WRITE;
/*!40000 ALTER TABLE `RegionTable` DISABLE KEYS */;
INSERT INTO `RegionTable` VALUES (1,1,0000010240,0000035200,0,0),(2,1,0000102400,0000352000,0,0);
/*!40000 ALTER TABLE `RegionTable` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2015-12-30 17:00:08
