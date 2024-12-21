-- MySQL dump 10.13  Distrib 8.0.39, for Win64 (x86_64)
--
-- Host: localhost    Database: personalbudgeting
-- ------------------------------------------------------
-- Server version	8.0.39

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `personalbudgeting`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `personalbudgeting` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

USE `personalbudgeting`;

--
-- Table structure for table `budget`
--

DROP TABLE IF EXISTS `budget`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `budget` (
  `budgetId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `userId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `checkingBalance` bigint NOT NULL,
  `savingsBalance` bigint DEFAULT '0',
  `budgetTotal` bigint NOT NULL,
  `budgetRemaining` bigint NOT NULL,
  `totalSpent` bigint DEFAULT '0',
  `updatedAt` datetime NOT NULL,
  `createdAt` datetime NOT NULL,
  PRIMARY KEY (`budgetId`),
  KEY `userId` (`userId`),
  CONSTRAINT `budget_ibfk_1` FOREIGN KEY (`userId`) REFERENCES `users` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `budget`
--

LOCK TABLES `budget` WRITE;
/*!40000 ALTER TABLE `budget` DISABLE KEYS */;
INSERT INTO `budget` VALUES ('21b18a16-f032-4479-8f36-8a7f357abdd2','07a1c65c-4c42-40fc-ba18-b50e20ca0c04',1900,0,1900,1900,7856,'2024-12-16 19:37:24','2024-12-16 17:05:09'),('5849c6b4-388f-47f3-b861-8bbf8ddc1a5b','70651e6b-3845-4097-9a20-1ad077f2f533',2000,2500,4500,4500,0,'2024-12-11 14:34:30','2024-12-11 14:31:16'),('ca0f3762-94d1-4fd5-b55f-718af388e2ca','1d4caa7f-81cb-46a5-a760-5479c1928d91',200,0,200,200,0,'2024-11-29 22:54:29','2024-11-29 22:51:36'),('d36260fa-a46f-4d2e-81aa-27e2c02106c6','c7395143-bbf9-4833-9e2d-3584f2f1dd72',200,200,400,400,0,'2024-11-29 21:52:12','2024-11-29 21:52:12'),('d8f01d49-c74d-49bd-bcd1-61b11d77388e','633b71d0-e439-4eca-8729-6258b90f8115',200,200,400,400,0,'2024-12-02 16:27:27','2024-12-02 16:27:27');
/*!40000 ALTER TABLE `budget` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `expensecategory`
--

DROP TABLE IF EXISTS `expensecategory`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `expensecategory` (
  `expenseCategoryId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `userId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `totalSum` bigint DEFAULT '0',
  PRIMARY KEY (`expenseCategoryId`),
  KEY `userId` (`userId`),
  CONSTRAINT `expensecategory_ibfk_1` FOREIGN KEY (`userId`) REFERENCES `users` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `expensecategory`
--

LOCK TABLES `expensecategory` WRITE;
/*!40000 ALTER TABLE `expensecategory` DISABLE KEYS */;
INSERT INTO `expensecategory` VALUES ('374d6ee0-fbe6-4c3a-acc5-025eaa243957','70651e6b-3845-4097-9a20-1ad077f2f533','new','',0),('9604afd0-191d-4823-a3cc-bdb75f50927e','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','statement','mandatory',456),('9d39bf00-23ed-4b39-8fa5-99216f0a1052','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','clothes','much needo',3300),('d4d341f3-7e54-4051-aef7-57e5bca6fdea','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','groceries','important stuff',4000),('d8ac3284-203b-4a94-afd1-b8be8fd0d210','70651e6b-3845-4097-9a20-1ad077f2f533','groceries','only important things',0);
/*!40000 ALTER TABLE `expensecategory` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `expenses`
--

DROP TABLE IF EXISTS `expenses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `expenses` (
  `expenseId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `userId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `categoryId` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `expenseType` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `amountInCents` bigint NOT NULL,
  `createdAt` datetime NOT NULL,
  PRIMARY KEY (`expenseId`),
  KEY `userId` (`userId`),
  KEY `expenses_ibfk_2` (`categoryId`),
  CONSTRAINT `expenses_ibfk_1` FOREIGN KEY (`userId`) REFERENCES `users` (`userId`),
  CONSTRAINT `expenses_ibfk_2` FOREIGN KEY (`categoryId`) REFERENCES `expensecategory` (`expenseCategoryId`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `expenses`
--

LOCK TABLES `expenses` WRITE;
/*!40000 ALTER TABLE `expenses` DISABLE KEYS */;
INSERT INTO `expenses` VALUES ('381425ac-ec3c-4fa5-9eee-e2dd5aa61900','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','9604afd0-191d-4823-a3cc-bdb75f50927e','','checkingBalance',333,'2024-12-16 19:36:46'),('456cfbb7-c528-47d4-9e4e-9ec131b1f0df','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','9604afd0-191d-4823-a3cc-bdb75f50927e','','checkingBalance',23,'2024-12-16 19:37:24'),('54805221-36ad-4ad6-9ada-eb69fae34722','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','9d39bf00-23ed-4b39-8fa5-99216f0a1052','new stuff','checkingBalance',3300,'2024-12-16 19:30:59'),('ad59457e-72a4-4fb3-866c-5ee9e9b4a437','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','d4d341f3-7e54-4051-aef7-57e5bca6fdea','','checkingBalance',4000,'2024-12-16 19:29:28'),('c44ed6fe-28a4-43e8-911d-b4b8853be3a3','07a1c65c-4c42-40fc-ba18-b50e20ca0c04','9604afd0-191d-4823-a3cc-bdb75f50927e','','checkingBalance',100,'2024-12-16 17:05:43');
/*!40000 ALTER TABLE `expenses` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sessions`
--

DROP TABLE IF EXISTS `sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sessions` (
  `token` char(43) COLLATE utf8mb4_unicode_ci NOT NULL,
  `data` blob NOT NULL,
  `expiry` timestamp(6) NOT NULL,
  PRIMARY KEY (`token`),
  KEY `sessions_expiry_idx` (`expiry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sessions`
--

LOCK TABLES `sessions` WRITE;
/*!40000 ALTER TABLE `sessions` DISABLE KEYS */;
INSERT INTO `sessions` VALUES ('-esGvkQYIthOqh5yxSbc9X2DSVsZWEXN-hDO1O8VXiw',_binary '%ˇÄ\0DeadlineˇÇ\0ValuesˇÑ\0\0\0ˇÅTimeˇÇ\0\0\0\'ˇÉmap[string]interface {}ˇÑ\0\0\0YˇÄ\0\0\0\ﬁ\ˆº1.¡€òˇˇauthenticatedUserIDstring&\0$07a1c65c-4c42-40fc-ba18-b50e20ca0c04\0','2024-12-20 06:15:29.784457');
/*!40000 ALTER TABLE `sessions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `userId` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hashedPassword` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `createdAt` datetime NOT NULL,
  `displayName` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`userId`),
  UNIQUE KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES ('07a1c65c-4c42-40fc-ba18-b50e20ca0c04','nika@gmail.com','$2a$12$Q.IUMitWUZAat5lJM/b32.77v1QWNvslZc0Wo7esAeTS1qgMYEKlm','2024-11-16 00:34:26','Nika'),('1d4caa7f-81cb-46a5-a760-5479c1928d91','phone@gmail.com','$2a$12$GaenC0XKGLR3BMBL86NEUON2g7qZwUeF56gxBVrA3XUOxQEICCRdm','2024-11-29 22:45:29','phone'),('24d238cc-1540-452c-8e96-14f779f370c2','noke@gmail.com','$2a$12$6xG0ndF.W9tW6NOafpEXDusEeDDe.OUjEDKsTtU758OgL6PXk/8GK','2024-11-26 13:39:24','noke'),('58d1bd13-814b-433e-9d4c-73e22372b79c','nike@gmail.com','$2a$12$687aPDzRTLfGbYwVE5ZYGuPDFbB2oLQKJJQV9J5w6ylYeij27XKFu','2024-11-26 14:30:10','nike@gmail.com'),('633b71d0-e439-4eca-8729-6258b90f8115','gavin@gmail.com','$2a$12$yOS9YCAPNM5DQWfPZOm1zOoj3LVd65HmgkMtTeYlV5ebyJW8h8B6.','2024-11-29 21:56:36','gavin'),('70651e6b-3845-4097-9a20-1ad077f2f533','creep@gmail.com','$2a$12$b622aC1DLmAUlgS7.LzAAuP/s.aqxSax6SdwIis5fcZQ.5jn.QYfO','2024-12-11 14:30:36','creep'),('c7395143-bbf9-4833-9e2d-3584f2f1dd72','nika@email.com','$2a$12$pUkeMP7CBdfHDiQZLg79q.076FsLdLGjCOmjSjiiQbKzRdTfeC5Zm','2024-11-29 16:36:50','Nika');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-12-21 10:18:29
