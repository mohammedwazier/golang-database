package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

const Version = "1.0.0"

type (
	Logger interface{
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warning(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct{
		mutex sync.Mutex
		mutexts map[string]*sync.Mutex
		dir string
		log Logger
	}
)
type Options struct{
	Logger
}

type Address struct{
	City string
	State string
	Country string
	Pincode json.Number
}

type User struct{
	Name string
	Age json.Number
	Contact string
	Company string
	Address Address
}

func New(dir string, options *Options)(*Driver, error){
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}
	if opts.Logger == nil {
		// opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
		// opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
		// log := lumber.NewConsoleLogger(lumber.INFO)
		// opts.Logger
	}

	driver := Driver {
		dir: dir,
		mutexts: make(map[string]*sync.Mutex),
		log: opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil{
		// opts.Logger.Debug(("Using '%s' (database already exist)\n"), dir)
		return &driver, nil
	}

	// opts.Logger.Debug("Creating the Database at '%s' ...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)


}

func (d *Driver) Write(collection, resource string, v interface{}) error{
	if collection == ""{
		return fmt.Errorf("Missing Collection - no Place to save the record!")
	}

	if resource == "" {
		return fmt.Errorf("Missing Resource - unable to Save the Record (No Name)!")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()

	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource + ".json")
	tmpPath := fnlPath + ".tmp"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tmpPath, b, 0655); err != nil {
		return err
	}

	return os.Rename(tmpPath, fnlPath)

}

func (d *Driver) Read(collection, resource string, v interface{}) error{
	if collection == ""{
		return fmt.Errorf("Missing Collection - no Place to save the record!")
	}

	if resource == "" {
		return fmt.Errorf("Missing Resource - unable to Save the Record (No Name)!")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string)([]string, error) {
	if collection == ""{
		return nil, fmt.Errorf("Missing Collection - Unable to Read!")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := ioutil.ReadDir(dir)

	var reacords []string

	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		reacords = append(reacords, string(b))
	}

	return reacords, nil
}

func (d *Driver) Delete(collection, resource string) error{
	path := filepath.Join(collection, resource)

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir);{
	case fi == nil, err != nil:
		return  fmt.Errorf("Unable to find file or directory named %v\n", path)

	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}
	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex{

	d.mutex.Lock()
	defer d.mutex.Unlock()

	m, ok := d.mutexts[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexts[collection] = m
	}

	return m
}

func stat(path string)(fi os.FileInfo, err error){
	if fi, err = os.Stat(path); os.IsNotExist(err){
		fi, err = os.Stat(path + ".json")
	}
	return
}

func main(){
	dir := "./"

	db, err := New(dir, nil)
	if err != nil{
		fmt.Println("Error", err)
	}

	employees := []User{
		{"John", "23", "Home", "Self", Address{"Karawang","West Java","Indonesia","41361"}},
		{"JohnA", "24", "Home", "Self", Address{"Karawang","West Java","Indonesia","41361"}},
		{"JohnB", "26", "Home", "Self", Address{"Karawang","West Java","Indonesia","41361"}},
		{"JohnC", "25", "Home", "Self", Address{"Karawang","West Java","Indonesia","41361"}},
		{"JohnD", "27", "Home", "Self", Address{"Karawang","West Java","Indonesia","41361"}},
	}

	for _, value := range employees {
		// First Parameter is Collection (Table)
		// Second Parameter is Value
		db.Write("users", value.Name, User{
			Name: value.Name,
			Age: value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	// Reads All Collections
	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error", err)
	}
	// Output Records
	fmt.Println(records)

	allUsers := []User{}

	for _, f := range records {
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil{
			fmt.Println("Error", err)
		}

		allUsers = append(allUsers, employeeFound)
	}

	fmt.Println(allUsers)

	if err := db.Delete("users", "john"); err != nil{
		fmt.Println("Error", err)
	}

	if err := db.Delete("users", ""); err != nil{
		fmt.Println("Error", err)
	}
}