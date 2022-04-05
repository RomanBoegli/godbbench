-- INIT
\benchmark once \name initialize
DROP SCHEMA IF EXISTS GoBench CASCADE; CREATE SCHEMA GoBench; CREATE TABLE GoBench.Customer (CustomerId INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50), Birthday  DATE); CREATE TABLE "gobench"."order" (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, CreationDate DATE, Comment VARCHAR(50), FOREIGN KEY (CustomerId) REFERENCES GoBench.Customer (CustomerId)); CREATE TABLE GoBench.Category (CategoryId INT PRIMARY KEY, Name VARCHAR(10)); CREATE TABLE GoBench.Supplier (SupplierId  INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50)); CREATE TABLE GoBench.Product (ProductId INT PRIMARY KEY, SupplierId INT NOT NULL, CategoryId INT NOT NULL, Code VARCHAR(6), Description VARCHAR(100), UnitSize INT, PricePerUnit DECIMAL(10,2), FOREIGN KEY (SupplierId) REFERENCES GoBench.Supplier (SupplierId), FOREIGN KEY (CategoryId)  REFERENCES GoBench.Category (CategoryId)); CREATE TABLE GoBench.LineItem (LineItemId INT PRIMARY KEY, OrderId INT NOT NULL, ProductId INT NOT NULL, Quantity INT, DeliveryDate DATE, FOREIGN KEY (OrderId) REFERENCES "gobench"."order" (OrderId), FOREIGN KEY (ProductId) REFERENCES GoBench.Product (ProductId));

-- INSERT
\benchmark loop 0.75 \name insert_customer
INSERT INTO GoBench.Customer (CustomerId, Name, Address, Birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');

\benchmark loop 1.0 \name insert_order
INSERT INTO GoBench.Order (OrderId, CustomerId, CreationDate, Comment) VALUES( {{.Iter}}, {{call .RandId "Customer" "Postgres" }}, '{{call .RandDate }}', '{{call .RandString 0 50 }}');

\benchmark loop 0.25 \name insert_supplier
INSERT INTO GoBench.Supplier (SupplierId, Name, Address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');

\benchmark loop 0.1 \name insert_category
INSERT INTO GoBench.Category (CategoryId, Name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');

\benchmark loop 0.5 \name insert_product
INSERT INTO GoBench.Product (ProductId, SupplierId, CategoryId, Code, Description, UnitSize, PricePerUnit) VALUES( {{.Iter}}, {{call .RandId "Supplier" "Postgres" }}, {{call .RandId "Category" "Postgres" }}, '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});

\benchmark loop 1.0 \name insert_lineitem
INSERT INTO GoBench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, {{call .RandId "Order" "Postgres" }}, {{call .RandId "Product" "Postgres" }}, {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');


-- SELECTS
\benchmark loop 1.0 \name select_simple
SELECT * FROM GoBench.Customer WHERE CustomerId = {{.Iter}} 

\benchmark loop 1.0 \name select_medium
SELECT * FROM GoBench.Product p JOIN GoBench.Supplier s on p.SupplierId = s.SupplierId WHERE s.SupplierId = {{.Iter}} ORDER BY p.PricePerUnit DESC
  
\benchmark loop 1.0 \name select_complex
SELECT c.CustomerId, c.Name, SUM(li.Quantity * p.UnitSize * p.PricePerUnit) as TotalOrderValue  FROM GoBench.Customer c  INNER JOIN GoBench.Order o on o.CustomerId = c.CustomerId  INNER JOIN GoBench.LineItem li on o.OrderId = li.OrderId  INNER JOIN GoBench.Product p on p.ProductId = li.ProductId  WHERE (o.CreationDate BETWEEN '{{call .RandDate }}' AND '9999-12-31')  GROUP by c.CustomerId, c.Name  ORDER by c.CustomerId




-- CLEAN
--\benchmark once \name clean
--DROP TABLE IF EXISTS GoBench.LineItem CASCADE; DROP TABLE IF EXISTS GoBench.Product CASCADE; DROP TABLE IF EXISTS GoBench.Category CASCADE; DROP TABLE IF EXISTS GoBench.Supplier CASCADE; DROP TABLE IF EXISTS GoBench.Order CASCADE; DROP TABLE IF EXISTS GoBench.Customer CASCADE; DROP SCHEMA IF EXISTS GoBench CASCADE;