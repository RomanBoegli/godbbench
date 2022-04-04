-- Create database & tables
\benchmark once \name initialize
DROP DATABASE IF EXISTS GoBench; CREATE DATABASE GoBench; USE GoBench; CREATE TABLE Customer (CustomerId INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50), Birthday  DATE);    CREATE TABLE `Order` (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, CreationDate DATE, Comment VARCHAR(50), FOREIGN KEY (CustomerId) REFERENCES Customer (CustomerId));    CREATE TABLE Category (CategoryId INT PRIMARY KEY, Name VARCHAR(10));    CREATE TABLE Supplier (SupplierId  INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50));    CREATE TABLE Product (ProductId INT PRIMARY KEY, SupplierId INT NOT NULL, CategoryId INT NOT NULL, Code VARCHAR(6), Description VARCHAR(100), UnitSize INT, PricePerUnit DECIMAL, FOREIGN KEY (SupplierId) REFERENCES Supplier (SupplierId), FOREIGN KEY (CategoryId) REFERENCES Category (CategoryId));    CREATE TABLE LineItem (LineItemId INT PRIMARY KEY, OrderId INT NOT NULL, ProductId INT NOT NULL, Quantity INT, DeliveryDate DATE, FOREIGN KEY (OrderId) REFERENCES `Order` (OrderId), FOREIGN KEY (ProductId) REFERENCES Product (ProductId));

-- INSERT
\benchmark loop 0.75 \name insert_customer
INSERT INTO GoBench.Customer (CustomerId, Name, Address, Birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');

\benchmark loop 1.0 \name insert_order
INSERT INTO GoBench.Order (OrderId, CustomerId, CreationDate, Comment) VALUES( {{.Iter}}, {{call .RandId "Customer" "MySQL" }}, '{{call .RandDate }}', '{{call .RandString 0 50 }}');

\benchmark loop 0.25 \name insert_customer
INSERT INTO GoBench.Supplier (SupplierId, Name, Address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');

\benchmark loop 0.1 \name insert_category
INSERT INTO GoBench.Category (CategoryId, Name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');

\benchmark loop 0.5 \name insert_product
INSERT INTO GoBench.Product (ProductId, SupplierId, CategoryId, Code, Description, UnitSize, PricePerUnit) VALUES( {{.Iter}}, {{call .RandId "Supplier" "MySQL" }}, {{call .RandId "Category" "MySQL" }}, '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});

\benchmark loop 1.0 \name insert_lineitem
INSERT INTO GoBench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, {{call .RandId "Order" "MySQL" }}, {{call .RandId "Product" "MySQL" }}, {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');


-- SELECTS
--\benchmark loop 1.0 \name select1
--INSERT INTO GoBench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, {{call .RandId "Order" "MySQL" }}, {{call .RandId "Product" "MySQL" }}, {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');



-- DROPS
--\benchmark once \name clean
--USE GoBench;
--SET FOREIGN_KEY_CHECKS=0;
--DROP TABLE GoBench.LineItem;
--DROP TABLE GoBench.Order;
--DROP TABLE GoBench.Customer;
--DROP TABLE GoBench.Product;
--DROP TABLE GoBench.Category;
--DROP TABLE GoBench.Supplier;
--SET FOREIGN_KEY_CHECKS=1;
--DROP DATABASE GoBench;