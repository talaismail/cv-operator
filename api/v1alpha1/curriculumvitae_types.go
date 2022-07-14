/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CurriculumVitaeSpec defines the desired state of CurriculumVitae
type CurriculumVitaeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name        Name        `json:"name"`
	ContactInfo ContactInfo `json:"contactInfo"`
	// +optional
	Experiences []Experience `json:"experiences,omitempty"`
	// +optional
	Studies []Study `json:"studies,omitempty"`
	// +optional
	Skills []Skill `json:"skills,omitempty"`
}

type Name struct {
	// Personal name of the applicant.
	FirstName string `json:"firstName"`
	// Middle name of the applicant.
	MiddleName string `json:"middleName"`
	// Family name of the applicant.
	LastName string `json:"lastName"`
}

type Experience struct {
	// Role at the company.
	Role string `json:"role"`
	// Starting date.
	Date string `json:"date"`
	// Company name.
	CompanyName string `json:"companyName"`
}

type Study struct {
	// Degree earned.
	Degree string `json:"degree"`
	// Start date.
	Date string `json:"date"`
	// Name of the university.
	UniversityName string `json:"universityName"`
	// Grade achieved.
	Grade string `json:"grade"`
}

type Skill struct {
	// Name of the skill.
	Name string `json:"name"`
	// Level of expertise.
	Level string `json:"level"`
	// Description of the skill.
	Description string `json:"description"`
}

type ContactInfo struct {
	// Postal code of the address.
	PostalCode string `json:"postalCode"`
	// Street of the address.
	Street string `json:"street"`
	// Country of residence.
	Country string `json:"country"`
	// Email address of the applicant.
	Email string `json:"email"`
	// Telephone number of the applicant.
	PhoneNumber string `json:"phoneNumber"`
}

// CurriculumVitaeStatus defines the observed state of CurriculumVitae
type CurriculumVitaeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CurriculumVitae is the Schema for the curriculumvitaes API
type CurriculumVitae struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CurriculumVitaeSpec   `json:"spec,omitempty"`
	Status CurriculumVitaeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CurriculumVitaeList contains a list of CurriculumVitae
type CurriculumVitaeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CurriculumVitae `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CurriculumVitae{}, &CurriculumVitaeList{})
}
