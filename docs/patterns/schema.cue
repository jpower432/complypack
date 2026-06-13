// SPDX-License-Identifier: Apache-2.0

// Schema for compliance pipeline pattern cards.
// Each card instantiates the pipeline pattern for a persona.
package patterns

#PatternCard: {
	persona: #Persona
	driver:  #Driver
	catalog: #Catalog
	governance: #Governance
	scopingModel: #ScopingModel
	applicability: #Applicability
	evaluation: #Evaluation
}

#Persona: {
	name: string
	description: string
}

#Driver: {
	name: string
	description: string
}

#Catalog: {
	name: string
	description: string
	examples?: [...string]
}

#Governance: {
	name: string
	description: string
}

#ScopingModel: {
	name: string
	description: string
	levels?: [...string]
}

#Evaluation: {
	name: string
	description: string
	examples?: [...string]
}

#Applicability: {
	name: string
	description: string
}
