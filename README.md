# Forcepoint-ngfw-licenses

## `config.yml` sample

`contact_info` section is mandatory

```yaml
---
concurrent_workers: 12     # Optional, default: 8
licences_output_dir: "out" # Optional, default: jar-files
resseller: ""              # Optional, default: ""

contact_info:
  firstname: "Foo"
  lastname:  "Bar"
  email:     "foo.bar@corp.com"
  phone:     "+33612345678"
  company:   "My Corp"
  address:   "12 rue Portalis"
  zip:       "75008"
  city:      "Paris"
  country:   "FR"
  state:     "75"

# Feature not yet implemented
#smc: # install licences on the SMC
#  ip: ""
#  port: ""
#  api_key: ""
```

## Usage

You have to download "Purchase" html files and drop them in the same directory than `forcepoint-licenses` binary.

### To verify PoS validity and status

This command will parse every html files and search for Forcepoint NGFW PoS. Each of them will be load on Forcepoint license center and registration status will be retrived.

```
> forcepoint-licenses verify
7 PoS read from 2 files

Found 7 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"VALID", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"VALID", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
```

### To register PoS

This command will `verify` all PoS, and register them, using informations from `config.yml` file.

```
> forcepoint-licenses register
7 PoS read from 2 files

Found 7 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"VALID", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"VALID", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}

2 new POS were registred

Found 7 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
```

### To download PoS

This command will `verify` and `register` all PoS using informations from `config.yml`, and then download licenses files.

```
> forcepoint-licenses download
7 PoS read from 2 files

Found 7 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}

0 new POS were registred

Found 7 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 1105 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 2101 Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}

7 license files were downloaded in './out/' directory
```

All steps can be done at once:

```
> forcepoint-licenses download
1 PoS read from 1 files

Found 1 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"VALID", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}

1 new POS were registred

Found 1 valid PoS:
- XXXXXXXXXX-XXXXXXXXXX {LicenseStatus:"REGISTERED", SN:"N0CXXXXXXXXX", ProductName:"Forcepoint NGFW 120W Appliance", Company:"My Corp"}

1 license files were downloaded in './out/' directory
```
